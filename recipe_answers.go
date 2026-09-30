package vergeos

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Question types the platform stores. Aliases below are folded onto these.
const (
	questionTypeBool     = "bool"
	questionTypeNum      = "num"
	questionTypeRAM      = "ram"
	questionTypeDiskSize = "disksize"
	questionTypeSeconds  = "seconds"
	questionTypeString   = "string"
	questionTypeHostname = "hostname"
	questionTypeTextArea = "textarea"
	questionTypePassword = "password"
	questionTypeNetwork  = "network"
	questionTypeList     = "list"
	questionTypeRow      = "row"
	questionTypeCluster  = "cluster"
)

// Words a bool question accepts, compared after trim and lower-case.
// Anything else is refused. The platform reads an unrecognized string as
// false, which is how a UEFI answer of "enabled" used to build a BIOS VM.
var (
	boolTrueWords  = []string{"true", "yes", "on", "1"}
	boolFalseWords = []string{"false", "no", "off", "0"}
)

// Stock Linux recipes publish a HOSTNAME pattern whose inner quantifier is
// one character too strict. The platform does not apply that pattern on
// deploy, so a two-letter name such as "db" succeeds there and would be the
// only name this check rejects. The corrected pattern is used for this one
// string. The error still quotes the pattern the recipe published.
const (
	stockHostnameRegex          = `[a-zA-Z]([a-zA-Z0-9_-]+[a-zA-Z0-9])?`
	stockHostnameRegexIntended  = `[a-zA-Z]([a-zA-Z0-9_-]*[a-zA-Z0-9])?`
	recipeSectionDatabase       = "$database"
	previewLogErrorAPI          = "Error executing API command"
	previewLogErrorRecipe       = "Error executing recipe"
	recipePreviewCompleteMarker = "Simulation complete"
)

func canonicalQuestionType(questionType string) string {
	switch strings.ToLower(strings.TrimSpace(questionType)) {
	case "boolean":
		return questionTypeBool
	case "number":
		return questionTypeNum
	case "disk_size", "disksize":
		return questionTypeDiskSize
	case "text_area", "textarea":
		return questionTypeTextArea
	case "row_selection":
		return questionTypeRow
	default:
		return strings.ToLower(strings.TrimSpace(questionType))
	}
}

func isNumericQuestion(questionType string) bool {
	switch questionType {
	case questionTypeNum, questionTypeRAM, questionTypeDiskSize, questionTypeSeconds:
		return true
	default:
		return false
	}
}

func isStringQuestion(questionType string) bool {
	switch questionType {
	case questionTypeString, questionTypeHostname, questionTypeTextArea, questionTypePassword:
		return true
	default:
		return false
	}
}

func isInternalQuestion(q RecipeQuestion) bool {
	switch canonicalQuestionType(q.Type) {
	case "hidden", "database_create", "database_edit", "database_find", "field":
		return true
	}
	return q.SectionName == recipeSectionDatabase
}

// recipeAnswersNeedNetworks reports whether resolving these answers has to
// list vnets. An integer network key does not. A name does, including a
// numeric name, because that name is checked before the text is read as a key.
func recipeAnswersNeedNetworks(questions []RecipeQuestion, answers RecipeAnswers) bool {
	if len(answers) == 0 {
		return false
	}
	for _, q := range questions {
		if canonicalQuestionType(q.Type) != questionTypeNetwork || q.Name == "" || isInternalQuestion(q) {
			continue
		}
		value, ok := answers[q.Name]
		if !ok {
			continue
		}
		text, ok := value.(string)
		if !ok {
			continue
		}
		if strings.TrimSpace(text) == RecipeNewInternalNetwork {
			continue
		}
		return true
	}
	return false
}

type networkNameHit struct {
	count int
	key   int
}

func indexNetworkNames(networks []Network) map[string]networkNameHit {
	index := make(map[string]networkNameHit, len(networks))
	for _, network := range networks {
		hit := index[network.Name]
		if hit.count == 0 {
			hit.key = int(network.Key)
		}
		hit.count++
		index[network.Name] = hit
	}
	return index
}

// resolveRecipeAnswers checks answers against questions and returns the
// document to send. The caller's map is not modified. Defaults are not
// copied into the result. A recipe with no questions is returned as a copy
// of answers, because there is nothing local to check them against.
func resolveRecipeAnswers(questions []RecipeQuestion, answers RecipeAnswers, networks []Network) (map[string]any, error) {
	if answers == nil {
		answers = RecipeAnswers{}
	}
	byName := make(map[string]RecipeQuestion, len(questions))
	for _, question := range questions {
		if question.Name == "" {
			continue
		}
		if _, exists := byName[question.Name]; !exists {
			byName[question.Name] = question
		}
	}

	resolved := make(map[string]any, len(answers))
	if len(byName) == 0 {
		for name, value := range answers {
			resolved[name] = value
		}
		return resolved, nil
	}

	var problems []string
	for name := range answers {
		if _, ok := byName[name]; !ok {
			problems = append(problems, fmt.Sprintf("unknown answer %q is not a question on this recipe", name))
		}
	}

	networksByName := indexNetworkNames(networks)
	for name, question := range byName {
		questionType := canonicalQuestionType(question.Type)
		value, supplied := answers[name]
		if isInternalQuestion(question) {
			if !supplied {
				continue
			}
			if reason := rejectContainer(name, value); reason != "" {
				problems = append(problems, reason)
				continue
			}
			resolved[name] = value
			continue
		}
		if !supplied {
			if question.Required && recipeDefaultEmpty(question.Default) {
				problems = append(problems, fmt.Sprintf("missing required answer %q (%s)%s",
					name, questionLabel(question), choiceSuffix(question)))
			}
			continue
		}
		if reason := rejectContainer(name, value); reason != "" {
			problems = append(problems, reason)
			continue
		}

		coerced, coerceErr := coerceAnswer(name, questionType, value)
		if coerceErr != "" {
			problems = append(problems, coerceErr)
			continue
		}
		if questionType != questionTypeBool && question.Required && answerEmpty(coerced) {
			problems = append(problems, fmt.Sprintf("required answer %q was supplied empty", name))
			continue
		}
		problems = append(problems, checkAnswerConstraints(name, question, questionType, coerced)...)
		if questionType == questionTypeNetwork {
			next, networkErr := resolveNetworkAnswer(name, coerced, networksByName)
			if networkErr != "" {
				problems = append(problems, networkErr)
				continue
			}
			coerced = next
		}
		if choiceErr := checkInlineChoices(name, question, questionType, value, coerced); choiceErr != "" {
			problems = append(problems, choiceErr)
			continue
		}
		resolved[name] = coerced
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		return nil, &ValidationError{Field: "answers", Message: strings.Join(problems, "; ")}
	}
	return resolved, nil
}

func questionLabel(question RecipeQuestion) string {
	if question.Display != "" {
		return question.Display
	}
	if question.Type != "" {
		return question.Type
	}
	return "question"
}

func choiceSuffix(question RecipeQuestion) string {
	if question.Table == "" && len(question.Choices) > 0 {
		return describeChoices(question.Choices)
	}
	if question.Table != "" {
		return fmt.Sprintf(", choices come from table %q", question.Table)
	}
	return ""
}

func describeChoices(choices RecipeChoices) string {
	if len(choices) == 0 {
		return ""
	}
	keys := make([]string, 0, len(choices))
	for key := range choices {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	limit := len(keys)
	ellipsis := ""
	if limit > 8 {
		limit = 8
		ellipsis = " ..."
	}
	parts := make([]string, limit)
	for i := 0; i < limit; i++ {
		parts[i] = keys[i] + "=" + choices[keys[i]]
	}
	return ", valid: " + strings.Join(parts, ", ") + ellipsis
}

func recipeDefaultEmpty(raw json.RawMessage) bool {
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "null" || text == `""` {
		return true
	}
	var flag bool
	if json.Unmarshal(raw, &flag) == nil {
		return !flag
	}
	var number float64
	if json.Unmarshal(raw, &number) == nil {
		return number == 0
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s == ""
	}
	return false
}

func answerEmpty(value any) bool {
	if value == nil {
		return true
	}
	text, ok := value.(string)
	return ok && text == ""
}

func rejectContainer(name string, value any) string {
	if value == nil {
		return ""
	}
	switch reflect.ValueOf(value).Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return fmt.Sprintf("answer %q must be a single value", name)
	default:
		return ""
	}
}

func coerceAnswer(name, questionType string, value any) (any, string) {
	if isNumericQuestion(questionType) {
		number, ok := wholeNumber(value)
		if !ok {
			return nil, fmt.Sprintf("answer %q must be a whole number for a %s question", name, questionType)
		}
		return number, ""
	}
	if questionType == questionTypeBool {
		flag, ok := coerceBool(value)
		if !ok {
			return nil, boolAnswerError(name, value)
		}
		return flag, ""
	}
	return value, ""
}

func coerceBool(value any) (bool, bool) {
	switch typed := value.(type) {
	case bool:
		return typed, true
	case int:
		return boolFromInt(int64(typed))
	case int8:
		return boolFromInt(int64(typed))
	case int16:
		return boolFromInt(int64(typed))
	case int32:
		return boolFromInt(int64(typed))
	case int64:
		return boolFromInt(typed)
	case uint:
		return boolFromInt(int64(typed))
	case uint8:
		return boolFromInt(int64(typed))
	case uint16:
		return boolFromInt(int64(typed))
	case uint32:
		return boolFromInt(int64(typed))
	case uint64:
		if typed > math.MaxInt64 {
			return false, false
		}
		return boolFromInt(int64(typed))
	case float32:
		return boolFromFloat(float64(typed))
	case float64:
		return boolFromFloat(typed)
	case json.Number:
		if typed == "0" || typed == "1" {
			return typed == "1", true
		}
		return false, false
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "yes", "on", "1":
			return true, true
		case "false", "no", "off", "0", "":
			return false, true
		default:
			return false, false
		}
	default:
		return false, false
	}
}

func boolFromInt(value int64) (bool, bool) {
	switch value {
	case 1:
		return true, true
	case 0:
		return false, true
	default:
		return false, false
	}
}

func boolFromFloat(value float64) (bool, bool) {
	if value == 1 {
		return true, true
	}
	if value == 0 {
		return false, true
	}
	return false, false
}

func boolAnswerError(name string, value any) string {
	return fmt.Sprintf("answer %q is not a recognized boolean (got %s); accepted values are %s and %s",
		name, formatAnswerValue(value), strings.Join(boolTrueWords, ", "), strings.Join(boolFalseWords, ", "))
}

func formatAnswerValue(value any) string {
	text, ok := value.(string)
	if ok {
		return strconv.Quote(text)
	}
	return fmt.Sprintf("%v", value)
}

func wholeNumber(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int8:
		return int64(typed), true
	case int16:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case uint:
		if uint64(typed) > math.MaxInt64 {
			return 0, false
		}
		return int64(typed), true
	case uint8:
		return int64(typed), true
	case uint16:
		return int64(typed), true
	case uint32:
		return int64(typed), true
	case uint64:
		if typed > math.MaxInt64 {
			return 0, false
		}
		return int64(typed), true
	case float32:
		return wholeFloat(float64(typed))
	case float64:
		return wholeFloat(typed)
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil
	case string:
		text := strings.TrimSpace(typed)
		if text == "" || strings.HasPrefix(text, "+") {
			return 0, false
		}
		parsed, err := strconv.ParseInt(text, 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func wholeFloat(value float64) (int64, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value != math.Trunc(value) {
		return 0, false
	}
	if value < math.MinInt64 || value > math.MaxInt64 {
		return 0, false
	}
	return int64(value), true
}

func checkAnswerConstraints(name string, question RecipeQuestion, questionType string, value any) []string {
	if isNumericQuestion(questionType) {
		return checkNumericConstraints(name, question, questionType, value)
	}
	text, ok := value.(string)
	if !ok || !isStringQuestion(questionType) {
		return nil
	}
	var problems []string
	length := int64(utf8.RuneCountInString(text))
	if question.Min.Set && length < question.Min.N {
		problems = append(problems, fmt.Sprintf("answer %q is %d characters; the recipe requires at least %d", name, length, question.Min.N))
	}
	if question.Max.Set && question.Max.N > 0 && length > question.Max.N {
		problems = append(problems, fmt.Sprintf("answer %q is %d characters; the recipe allows at most %d", name, length, question.Max.N))
	}
	if question.Regex != "" {
		matched, err := regexFullMatch(regexToApply(question.Regex), text)
		if err == nil && !matched {
			problems = append(problems, fmt.Sprintf("answer %q does not match the recipe's pattern %s", name, question.Regex))
		}
	}
	return problems
}

func checkNumericConstraints(name string, question RecipeQuestion, questionType string, value any) []string {
	number, ok := value.(int64)
	if !ok {
		return []string{fmt.Sprintf("answer %q must be a whole number for a %s question", name, questionType)}
	}
	var problems []string
	floor := int64(0)
	floorLabel := fmt.Sprintf("a %s cannot be negative", questionType)
	if question.Min.Set {
		floor = question.Min.N
		floorLabel = fmt.Sprintf("the recipe requires at least %d", question.Min.N)
	}
	if number < floor {
		problems = append(problems, fmt.Sprintf("answer %q is %d; %s", name, number, floorLabel))
	}
	if question.Max.Set && question.Max.N > 0 && number > question.Max.N {
		problems = append(problems, fmt.Sprintf("answer %q is %d; the recipe allows at most %d", name, number, question.Max.N))
	}
	if questionType == questionTypeDiskSize && number > 0 && number < RecipeDiskSizeMinBytes {
		problems = append(problems, fmt.Sprintf(
			"answer %q is %d; a disk size answer is in bytes, so 50 GB is %d, and a value above zero but under 1 MB (%d) is refused",
			name, number, RecipeDiskSize50GB, RecipeDiskSizeMinBytes))
	}
	return problems
}

func regexToApply(pattern string) string {
	if pattern == stockHostnameRegex {
		return stockHostnameRegexIntended
	}
	return pattern
}

func regexFullMatch(pattern, value string) (bool, error) {
	compiled, err := regexp.Compile(`\A(?:` + pattern + `)\z`)
	if err != nil {
		return false, err
	}
	return compiled.MatchString(value), nil
}

func resolveNetworkAnswer(name string, value any, networks map[string]networkNameHit) (any, string) {
	text, ok := value.(string)
	if !ok {
		number, isNumber := wholeNumber(value)
		if !isNumber {
			return nil, fmt.Sprintf("answer %q must be a vnet key or a network name", name)
		}
		return number, ""
	}
	text = strings.TrimSpace(text)
	if text == RecipeNewInternalNetwork {
		return text, ""
	}
	hit, found := networks[text]
	if found && hit.count > 1 {
		return nil, fmt.Sprintf("network name %q is ambiguous, %d networks share it; give the vnet key instead (answer %q)", text, hit.count, name)
	}
	if found && hit.count == 1 {
		return int64(hit.key), ""
	}
	if number, ok := wholeNumber(text); ok {
		return number, ""
	}
	return nil, fmt.Sprintf("network %q not found (answer %q)", text, name)
}

func checkInlineChoices(name string, question RecipeQuestion, questionType string, original, coerced any) string {
	if questionType != questionTypeList || question.Table != "" || len(question.Choices) == 0 {
		return ""
	}
	text := choiceText(coerced)
	if _, ok := question.Choices[text]; ok {
		return ""
	}
	hint := ""
	if _, isBool := original.(bool); !isBool {
		hint = listLabelHint(text, question.Choices)
	}
	return fmt.Sprintf("answer %q is not a valid choice (got %s)%s%s%s",
		name, formatAnswerValue(original), describeChoices(question.Choices), hint, yamlBoolListHint(original, question.Choices))
}

func choiceText(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case bool:
		if typed {
			return "True"
		}
		return "False"
	case int64:
		return strconv.FormatInt(typed, 10)
	case int:
		return strconv.Itoa(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func listLabelHint(value string, choices RecipeChoices) string {
	folded := strings.ToLower(strings.TrimSpace(value))
	if folded == "" {
		return ""
	}
	var labelHits []string
	for key, label := range choices {
		if strings.ToLower(strings.TrimSpace(label)) == folded {
			labelHits = append(labelHits, key)
		}
	}
	sort.Strings(labelHits)
	switch len(labelHits) {
	case 1:
		return fmt.Sprintf("; %q is the label for key %q", value, labelHits[0])
	default:
		if len(labelHits) > 1 {
			quoted := make([]string, len(labelHits))
			for i, key := range labelHits {
				quoted[i] = strconv.Quote(key)
			}
			return fmt.Sprintf("; %q is the label for keys %s", value, strings.Join(quoted, ", "))
		}
	}
	var keyHits []string
	for key := range choices {
		if strings.ToLower(key) == folded && key != value {
			keyHits = append(keyHits, key)
		}
	}
	sort.Strings(keyHits)
	if len(keyHits) == 1 {
		return fmt.Sprintf("; %q matches key %q", value, keyHits[0])
	}
	return ""
}

func yamlBoolListHint(value any, choices RecipeChoices) string {
	flag, ok := value.(bool)
	if !ok {
		return ""
	}
	byLower := make(map[string]string, len(choices))
	for key := range choices {
		lower := strings.ToLower(key)
		if _, exists := byLower[lower]; !exists {
			byLower[lower] = key
		}
	}
	words := []string{"yes", "true", "on", "no", "false", "off"}
	present := false
	for _, word := range words {
		if _, ok := byLower[word]; ok {
			present = true
			break
		}
	}
	if !present {
		return ""
	}
	preferred := []string{"yes", "true", "on"}
	if !flag {
		preferred = []string{"no", "false", "off"}
	}
	var example string
	for _, word := range append(preferred, words...) {
		if key, ok := byLower[word]; ok {
			example = key
			break
		}
	}
	if example == "" {
		return ""
	}
	return fmt.Sprintf("; a boolean is not a choice key; quote the string (for example %q)", example)
}

func previewLogErrors(logs []string) []string {
	var found []string
	for _, line := range logs {
		if strings.Contains(line, previewLogErrorAPI) || strings.Contains(line, previewLogErrorRecipe) {
			found = append(found, line)
		}
	}
	return found
}
