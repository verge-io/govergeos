package vergeos

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// RecipeNewInternalNetwork is the network answer that tells a recipe to
// create a new internal network. It is sent unchanged. Any other network
// name is resolved to a vnet key before the request is sent.
const RecipeNewInternalNetwork = "__new_internal__"

// Disk-size answers are bytes. Zero means "use the recipe default". A value
// above zero and under 1 MB is refused: 50 is fifty bytes, and the platform
// then keeps the image's own size while reporting success.
const (
	// RecipeDiskSizeMinBytes is the smallest disk size, other than zero, that
	// a disksize answer may use.
	RecipeDiskSizeMinBytes int64 = 1 << 20
	// RecipeDiskSize50GB is 50 gigabytes in bytes. Refusals quote it so a
	// gigabyte-shaped answer has the byte form next to it.
	RecipeDiskSize50GB int64 = 50 * (1 << 30)
)

// RecipeAnswers is the caller's answer to a recipe's questions, keyed by
// question name. Deploy and Preview check each value against that question's
// type before anything is sent.
type RecipeAnswers map[string]any

// RecipeAnswerSet is the answer document stored on a recipe instance.
// The platform may return it as an object or as a string containing one.
type RecipeAnswerSet map[string]any

// UnmarshalJSON accepts a JSON object or a string that contains one.
func (s *RecipeAnswerSet) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "" || text == "null" {
		return nil
	}
	if text[0] == '"' {
		var encoded string
		if err := json.Unmarshal(data, &encoded); err != nil {
			return err
		}
		encoded = strings.TrimSpace(encoded)
		if encoded == "" || encoded == "null" {
			return nil
		}
		data = []byte(encoded)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*s = decoded
	return nil
}

// RecipeBound is a question's min or max.
//
// Set is false when the question did not declare the bound, including when
// the API sends null or "". A set zero is a real value: max 0 means there is
// no upper limit, and min 0 is a floor of zero.
type RecipeBound struct {
	Set bool
	N   int64
}

// UnmarshalJSON accepts a JSON number or a numeric string.
func (b *RecipeBound) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "" || text == "null" || text == `""` {
		*b = RecipeBound{}
		return nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*b = RecipeBound{Set: true, N: n}
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		s = strings.TrimSpace(s)
		if s == "" {
			*b = RecipeBound{}
			return nil
		}
		parsed, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("vergeos: cannot unmarshal %s into RecipeBound", text)
		}
		*b = RecipeBound{Set: true, N: parsed}
		return nil
	}
	var f float64
	if err := json.Unmarshal(data, &f); err == nil && f == float64(int64(f)) {
		*b = RecipeBound{Set: true, N: int64(f)}
		return nil
	}
	return fmt.Errorf("vergeos: cannot unmarshal %s into RecipeBound", text)
}

// MarshalJSON writes the bound as a JSON number, or null when it was not set.
func (b RecipeBound) MarshalJSON() ([]byte, error) {
	if !b.Set {
		return []byte("null"), nil
	}
	return json.Marshal(b.N)
}

// RecipeChoices is the fixed key/label map published on a list question.
// A table-backed list leaves this empty; its rows come from Table.
// The platform may send the map as an object or as a JSON string.
type RecipeChoices map[string]string

// UnmarshalJSON accepts an object, a JSON string containing an object, or
// an empty value. A shape that is not a choice map is ignored so one odd
// question cannot fail the whole list.
func (c *RecipeChoices) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "" || text == "null" || text == `""` || text == "[]" {
		return nil
	}
	if text[0] == '"' {
		var encoded string
		if err := json.Unmarshal(data, &encoded); err != nil {
			return err
		}
		encoded = strings.TrimSpace(encoded)
		if encoded == "" || encoded == "null" || encoded == "[]" {
			return nil
		}
		data = []byte(encoded)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}
	out := make(RecipeChoices, len(raw))
	for key, value := range raw {
		if key == "" || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			out[key] = typed
		default:
			out[key] = fmt.Sprint(typed)
		}
	}
	*c = out
	return nil
}

// VMRecipe is a template that deploys a virtual machine.
//
// Recipe keys are 40-character hex strings, the same value as ID.
type VMRecipe struct {
	// Key is the recipe's primary key.
	Key string `json:"$key,omitempty"`
	// ID is the same hex string as Key.
	ID string `json:"id,omitempty"`
	// Name is the recipe name.
	Name string `json:"name,omitempty"`
	// Description is the recipe description.
	Description string `json:"description,omitempty"`
	// Icon is the UI icon name.
	Icon string `json:"icon,omitempty"`
	// Version is the recipe version string.
	Version string `json:"version,omitempty"`
	// Build is the recipe build number.
	Build FlexInt `json:"build,omitempty"`
	// Catalog is the parent catalog key.
	Catalog string `json:"catalog,omitempty"`
	// CatalogDisplay is the parent catalog's display label.
	CatalogDisplay string `json:"catalog_display,omitempty"`
	// CatalogName is the parent catalog's name.
	CatalogName string `json:"catalog_name,omitempty"`
	// CatalogRepository is the parent catalog's repository key.
	CatalogRepository FlexInt `json:"catalog_repository,omitempty"`
	// Downloaded reports whether the recipe has been downloaded and can be deployed.
	// A list filter for this flag is "downloaded eq 1" or "downloaded eq 0".
	Downloaded bool `json:"downloaded,omitempty"`
	// UpdateAvailable reports whether the catalog has a newer build.
	UpdateAvailable bool `json:"update_available,omitempty"`
	// NeedsRepublish reports whether a local recipe has unpublished edits.
	NeedsRepublish bool `json:"needs_republish,omitempty"`
	// VM is the golden-image VM key for a local recipe. It is nil when the
	// recipe has no local VM.
	VM *FlexInt `json:"vm,omitempty"`
	// VMSnapshot is the golden-image snapshot key.
	VMSnapshot *FlexInt `json:"vm_snapshot,omitempty"`
	// Creator is the user who created the recipe.
	Creator string `json:"creator,omitempty"`
}

// RecipeQuestion is one input on a recipe's deploy form.
type RecipeQuestion struct {
	// Key is the question row id.
	Key FlexInt `json:"$key,omitempty"`
	// Recipe is the owning recipe reference, such as "vm_recipes/{key}".
	Recipe string `json:"recipe,omitempty"`
	// Section is the section row id.
	Section FlexInt `json:"section,omitempty"`
	// SectionName is the section name. "$database" marks recipe plumbing
	// rather than an operator question.
	SectionName string `json:"section_name,omitempty"`
	// Name is the answer key, and the variable name the recipe script reads.
	Name string `json:"name,omitempty"`
	// Display is the label shown on the form.
	Display string `json:"display,omitempty"`
	// Hint is placeholder text.
	Hint string `json:"hint,omitempty"`
	// Help is the tooltip.
	Help string `json:"help,omitempty"`
	// Note is the text shown under the field.
	Note string `json:"note,omitempty"`
	// Type is the question type (string, bool, num, disksize, network, list, and so on).
	Type string `json:"type,omitempty"`
	// Default is the platform's default. Null, "", 0, and false count as
	// empty when a required question was left unanswered. The platform applies
	// defaults itself; Deploy does not send them back.
	Default json.RawMessage `json:"default,omitempty"`
	// Required reports whether the question must be answered.
	Required bool `json:"required,omitempty"`
	// Enabled reports whether the question is on the form.
	Enabled bool `json:"enabled,omitempty"`
	// ReadOnly reports whether the question is read-only after creation.
	ReadOnly bool `json:"readonly,omitempty"`
	// OrderID is the display order inside the section.
	OrderID FlexInt `json:"orderid,omitempty"`
	// Min is the declared minimum. For a number it is a value floor. For a
	// string it is a length floor. An unset bound is not the same as zero.
	Min RecipeBound `json:"min,omitempty"`
	// Max is the declared maximum. Zero means there is no upper limit.
	Max RecipeBound `json:"max,omitempty"`
	// Regex is the pattern the whole answer must match, when the question declares one.
	Regex string `json:"regex,omitempty"`
	// Choices is the inline list of accepted keys. Empty for a table-backed list.
	Choices RecipeChoices `json:"list,omitempty"`
	// Table is the database table a row, list, or cluster question reads.
	Table string `json:"table,omitempty"`
	// Fields is the field list used to render that table.
	Fields string `json:"fields,omitempty"`
	// Filter is the table filter published on the question.
	Filter string `json:"filter,omitempty"`
	// DatabaseContext selects the parent database or the new tenant's database.
	DatabaseContext string `json:"database_context,omitempty"`
	// HideNone reports whether the form hides an empty list option.
	HideNone bool `json:"hide_none,omitempty"`
	// Conditions is the question's visibility rules, kept as published.
	Conditions json.RawMessage `json:"conditions,omitempty"`
	// OnChange is the question's change handler, kept as published.
	OnChange json.RawMessage `json:"on_change,omitempty"`
	// PostprocessString is the platform's post-processing hint.
	PostprocessString string `json:"postprocess_string,omitempty"`
	// DontStore reports whether the answer is kept off the instance row.
	// Passwords are often marked this way and are still sent on deploy.
	DontStore bool `json:"dont_store,omitempty"`
	// System reports whether the question is platform-owned.
	System bool `json:"system,omitempty"`
}

// VMRecipeDeployRequest is the input to Deploy and to Preview.
//
// Preview takes the same request and does not deploy it. There is no flag
// on this struct that selects a dry run.
type VMRecipeDeployRequest struct {
	// Recipe is the recipe key (40-character hex).
	Recipe string `json:"recipe"`
	// Name is the name of the VM to create.
	Name string `json:"name"`
	// Answers are checked against the recipe's questions before the request
	// is sent. Nil is an empty set. Keys the recipe does not define are refused
	// when the recipe publishes questions.
	Answers RecipeAnswers `json:"answers,omitempty"`
	// AutoUpdate asks the platform to move the VM when the recipe is updated.
	// Nil leaves the platform default in place.
	AutoUpdate *bool `json:"auto_update,omitempty"`
}

// VMRecipeInstance is one VM deployed from a recipe.
type VMRecipeInstance struct {
	// Key is the instance row id.
	Key FlexInt `json:"$key,omitempty"`
	// Recipe is the recipe key this instance was deployed from.
	Recipe string `json:"recipe,omitempty"`
	// RecipeName is that recipe's name.
	RecipeName string `json:"recipe_name,omitempty"`
	// VM is the VM that was created.
	VM FlexInt `json:"vm,omitempty"`
	// VMName is that VM's name.
	VMName string `json:"vm_name,omitempty"`
	// Name is the instance name.
	Name string `json:"name,omitempty"`
	// Version is the recipe version at deploy time.
	Version string `json:"version,omitempty"`
	// Build is the recipe build at deploy time.
	Build FlexInt `json:"build,omitempty"`
	// AutoUpdate reports whether the instance follows recipe updates.
	AutoUpdate bool `json:"auto_update,omitempty"`
	// Created is the creation time in microseconds.
	Created int64 `json:"created,omitempty"`
	// Modified is the last modification time in microseconds.
	Modified int64 `json:"modified,omitempty"`
	// Answers is the stored answer document. Get includes it. List does not,
	// because the document can contain a guest password.
	Answers RecipeAnswerSet `json:"answers,omitempty"`
}

// VMRecipeCloudInitFile is one file rendered by a recipe preview.
// Contents may include guest credentials.
type VMRecipeCloudInitFile struct {
	// Name is the file name, such as "user-data".
	Name string `json:"name,omitempty"`
	// Contents is the rendered file body.
	Contents string `json:"contents,omitempty"`
}

// VMRecipePreview is the report from a recipe dry run.
//
// Resource keys inside Answers are predictions. Nothing in this report is a
// VMRecipeInstance, and Preview does not look one up.
type VMRecipePreview struct {
	// CloudInitFiles are the rendered guest-configuration files.
	CloudInitFiles []VMRecipeCloudInitFile `json:"cloudinit_files"`
	// Logs are the steps the platform walked.
	Logs []string `json:"logs"`
	// Answers are the values the platform resolved, including predicted keys.
	Answers map[string]any `json:"answers"`
}

const (
	vmRecipeFields = "$key,id,name,description,icon,version,build,catalog," +
		"catalog#$display as catalog_display,catalog#name as catalog_name," +
		"catalog#repository as catalog_repository,downloaded,update_available," +
		"needs_republish,vm,vm_snapshot,creator"

	recipeQuestionFields = "$key,recipe,section,section#name as section_name,name,display,hint,help,note," +
		"type,default,required,enabled,readonly,orderid,min,max,regex,list,table,fields,filter," +
		"database_context,hide_none,conditions,on_change,postprocess_string,dont_store,system"

	vmRecipeInstanceListFields = "$key,recipe,recipe#name as recipe_name,vm,vm#name as vm_name,name,version,build,auto_update,created,modified"

	vmRecipeInstanceGetFields = vmRecipeInstanceListFields + ",answers"
)
