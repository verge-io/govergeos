package vergeos

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestResolveRecipeAnswers_BoolRefusesUnknown(t *testing.T) {
	questions := []RecipeQuestion{{
		Name: "SELECT_CREATE_UEFI", Type: "bool", Display: "UEFI",
	}}
	accepted := []struct {
		value any
		want  bool
	}{
		{true, true}, {false, false},
		{"true", true}, {"yes", true}, {"on", true}, {"1", true},
		{"false", false}, {"no", false}, {"off", false}, {"0", false}, {"", false},
		{1, true}, {0, false}, {int64(1), true}, {float64(0), false},
	}
	for _, tc := range accepted {
		got, err := resolveRecipeAnswers(questions, RecipeAnswers{"SELECT_CREATE_UEFI": tc.value}, nil)
		if err != nil {
			t.Fatalf("value %#v: %v", tc.value, err)
		}
		flag, ok := got["SELECT_CREATE_UEFI"].(bool)
		if !ok || flag != tc.want {
			t.Fatalf("value %#v resolved to %#v, want %v", tc.value, got["SELECT_CREATE_UEFI"], tc.want)
		}
	}

	for _, value := range []any{"enabled", "UEFI", "maybe", 2, 1.5} {
		_, err := resolveRecipeAnswers(questions, RecipeAnswers{"SELECT_CREATE_UEFI": value}, nil)
		if !IsValidationError(err) {
			t.Fatalf("value %#v err = %v, want validation", value, err)
		}
		message := err.Error()
		if strings.Contains(strings.ToLower(message), "treated as false") {
			t.Fatalf("value %#v was described as false: %s", value, message)
		}
		if !strings.Contains(message, "not a recognized boolean") || !strings.Contains(message, "true, yes, on, 1") {
			t.Fatalf("value %#v message = %s", value, message)
		}
	}
}

func TestResolveRecipeAnswers_DiskSizeIsBytes(t *testing.T) {
	questions := []RecipeQuestion{{Name: "YB_DRIVE_OS_SIZE", Type: "disksize"}}

	_, err := resolveRecipeAnswers(questions, RecipeAnswers{"YB_DRIVE_OS_SIZE": 50}, nil)
	if !IsValidationError(err) {
		t.Fatalf("50: %v", err)
	}
	message := err.Error()
	for _, part := range []string{"bytes", "53687091200", "1048576"} {
		if !strings.Contains(message, part) {
			t.Fatalf("message %q missing %s", message, part)
		}
	}

	_, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_DRIVE_OS_SIZE": "50"}, nil)
	if !IsValidationError(err) || !strings.Contains(err.Error(), "bytes") {
		t.Fatalf("string 50: %v", err)
	}
	_, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_DRIVE_OS_SIZE": int64(RecipeDiskSizeMinBytes - 1)}, nil)
	if err == nil {
		t.Fatal("1 MB minus one byte was accepted")
	}

	for _, value := range []any{0, int64(RecipeDiskSizeMinBytes), int64(RecipeDiskSize50GB), "21474836480"} {
		got, err := resolveRecipeAnswers(questions, RecipeAnswers{"YB_DRIVE_OS_SIZE": value}, nil)
		if err != nil {
			t.Fatalf("value %#v: %v", value, err)
		}
		if _, ok := got["YB_DRIVE_OS_SIZE"].(int64); !ok {
			t.Fatalf("value %#v resolved to %#v", value, got["YB_DRIVE_OS_SIZE"])
		}
	}

	_, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_DRIVE_OS_SIZE": "4096MB"}, nil)
	if err == nil || !strings.Contains(err.Error(), "whole number") {
		t.Fatalf("4096MB: %v", err)
	}
	_, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_DRIVE_OS_SIZE": true}, nil)
	if err == nil || !strings.Contains(err.Error(), "whole number") {
		t.Fatalf("bool disk: %v", err)
	}
	_, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_DRIVE_OS_SIZE": -1}, nil)
	if err == nil || !strings.Contains(err.Error(), "cannot be negative") {
		t.Fatalf("negative: %v", err)
	}
}

func TestResolveRecipeAnswers_RequiredDefaultsAndUnknown(t *testing.T) {
	questions := []RecipeQuestion{
		{Name: "HOSTNAME", Type: "string", Display: "Hostname", Required: true},
		{Name: "USER", Type: "string", Required: true, Default: json.RawMessage(`"ops"`)},
		{Name: "CORES", Type: "num", Required: true, Default: json.RawMessage(`0`)},
		{Name: "TIER", Type: "string", Required: true, Default: json.RawMessage(`"0"`)},
		{Name: "NOTE", Type: "string"},
	}
	_, err := resolveRecipeAnswers(questions, RecipeAnswers{"HOSTNAME": "web"}, nil)
	if err == nil || !strings.Contains(err.Error(), `missing required answer "CORES"`) {
		t.Fatalf("numeric zero default: %v", err)
	}
	if strings.Contains(err.Error(), "USER") || strings.Contains(err.Error(), "TIER") || strings.Contains(err.Error(), "NOTE") {
		t.Fatalf("optional or defaulted questions were required: %v", err)
	}

	got, err := resolveRecipeAnswers(questions, RecipeAnswers{"HOSTNAME": "web", "CORES": 2}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, sent := got["USER"]; sent {
		t.Fatal("defaulted USER was sent")
	}
	if got["CORES"] != int64(2) || got["HOSTNAME"] != "web" {
		t.Fatalf("resolved = %#v", got)
	}

	_, err = resolveRecipeAnswers(questions, RecipeAnswers{"HOSTNAME": "web", "CORES": 2, "EXTRA": 1}, nil)
	if err == nil || !strings.Contains(err.Error(), `unknown answer "EXTRA"`) {
		t.Fatalf("unknown: %v", err)
	}

	got, err = resolveRecipeAnswers(nil, RecipeAnswers{"SELECT_CREATE_UEFI": "enabled"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got["SELECT_CREATE_UEFI"] != "enabled" {
		t.Fatalf("passthrough = %#v", got)
	}
}

func TestResolveRecipeAnswers_ConstraintsAndList(t *testing.T) {
	questions := []RecipeQuestion{
		{Name: "HOSTNAME", Type: "hostname", Regex: stockHostnameRegex, Max: RecipeBound{Set: true, N: 0}},
		{Name: "YB_CPU_CORES", Type: "num", Min: RecipeBound{Set: true, N: 1}, Max: RecipeBound{Set: true, N: 8}},
		{Name: "YB_IP_ADDR_TYPE", Type: "list", Choices: RecipeChoices{"dhcp": "DHCP", "static": "Static"}},
		{Name: "PASSWORD", Type: "password", Min: RecipeBound{Set: true, N: 8}},
		{Name: "BROKEN", Type: "string", Regex: "("},
	}
	got, err := resolveRecipeAnswers(questions, RecipeAnswers{
		"HOSTNAME":        "db",
		"YB_CPU_CORES":    2,
		"YB_IP_ADDR_TYPE": "dhcp",
		"PASSWORD":        "correct-horse",
		"BROKEN":          "anything",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got["HOSTNAME"] != "db" {
		t.Fatalf("hostname = %#v", got["HOSTNAME"])
	}

	_, err = resolveRecipeAnswers(questions, RecipeAnswers{
		"HOSTNAME": "1bad", "YB_CPU_CORES": 9, "YB_IP_ADDR_TYPE": "DHCP", "PASSWORD": "short",
	}, nil)
	if err == nil {
		t.Fatal("expected constraint errors")
	}
	message := err.Error()
	for _, part := range []string{"pattern", "at most 8", "label for key", "at least 8"} {
		if !strings.Contains(message, part) {
			t.Fatalf("message %q missing %s", message, part)
		}
	}
	if strings.Contains(message, "short") || strings.Contains(message, "correct-horse") {
		t.Fatalf("password leaked into %s", message)
	}
}

func TestResolveRecipeAnswers_NetworkNames(t *testing.T) {
	questions := []RecipeQuestion{{Name: "YB_NIC_ETH0", Type: "network"}}
	networks := []Network{
		{Key: 12, Name: "Internal"},
		{Key: 3, Name: "3"},
		{Key: 4, Name: "Dup"},
		{Key: 5, Name: "Dup"},
	}

	got, err := resolveRecipeAnswers(questions, RecipeAnswers{"YB_NIC_ETH0": "Internal"}, networks)
	if err != nil || got["YB_NIC_ETH0"] != int64(12) {
		t.Fatalf("name: %#v %v", got, err)
	}
	got, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_NIC_ETH0": "3"}, networks)
	if err != nil || got["YB_NIC_ETH0"] != int64(3) {
		t.Fatalf("numeric name: %#v %v", got, err)
	}
	got, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_NIC_ETH0": "9"}, networks)
	if err != nil || got["YB_NIC_ETH0"] != int64(9) {
		t.Fatalf("numeric key: %#v %v", got, err)
	}
	got, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_NIC_ETH0": RecipeNewInternalNetwork}, networks)
	if err != nil || got["YB_NIC_ETH0"] != RecipeNewInternalNetwork {
		t.Fatalf("new internal: %#v %v", got, err)
	}
	got, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_NIC_ETH0": 12}, nil)
	if err != nil || got["YB_NIC_ETH0"] != int64(12) {
		t.Fatalf("int key: %#v %v", got, err)
	}

	_, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_NIC_ETH0": "Dup"}, networks)
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("ambiguous: %v", err)
	}
	_, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_NIC_ETH0": "Missing"}, networks)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("missing: %v", err)
	}
	if recipeAnswersNeedNetworks(questions, RecipeAnswers{"YB_NIC_ETH0": 12}) {
		t.Fatal("integer network key asked for a network list")
	}
	if !recipeAnswersNeedNetworks(questions, RecipeAnswers{"YB_NIC_ETH0": "3"}) {
		t.Fatal("numeric network name skipped the network list")
	}
}

func TestResolveRecipeAnswers_InternalAndContainer(t *testing.T) {
	questions := []RecipeQuestion{
		{Name: "HIDDEN", Type: "hidden", Required: true},
		{Name: "PLUMB", Type: "string", SectionName: "$database", Required: true},
		{Name: "HOSTNAME", Type: "string", Required: true},
	}
	got, err := resolveRecipeAnswers(questions, RecipeAnswers{"HOSTNAME": "web", "HIDDEN": "keep"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got["HIDDEN"] != "keep" {
		t.Fatalf("hidden = %#v", got["HIDDEN"])
	}
	if _, sent := got["PLUMB"]; sent {
		t.Fatal("unanswered internal question was required")
	}
	_, err = resolveRecipeAnswers(questions, RecipeAnswers{"HOSTNAME": map[string]any{"a": 1}}, nil)
	if err == nil || !strings.Contains(err.Error(), "single value") {
		t.Fatalf("container: %v", err)
	}
}

func TestResolveRecipeAnswers_FloatCountsAsWholeNumber(t *testing.T) {
	questions := []RecipeQuestion{{Name: "YB_RAM", Type: "ram"}, {Name: "YB_DRIVE_OS_SIZE", Type: "disksize"}}
	got, err := resolveRecipeAnswers(questions, RecipeAnswers{
		"YB_RAM":           float64(4096),
		"YB_DRIVE_OS_SIZE": float64(RecipeDiskSize50GB),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got["YB_RAM"] != int64(4096) || got["YB_DRIVE_OS_SIZE"] != int64(RecipeDiskSize50GB) {
		t.Fatalf("resolved = %#v", got)
	}
	_, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_RAM": 1.5}, nil)
	if err == nil || !strings.Contains(err.Error(), "whole number") {
		t.Fatalf("fraction: %v", err)
	}
	_, err = resolveRecipeAnswers(questions, RecipeAnswers{"YB_DRIVE_OS_SIZE": float64(50)}, nil)
	if err == nil || !strings.Contains(err.Error(), "bytes") {
		t.Fatalf("float 50: %v", err)
	}
}

func TestRecipeAnswerSet_UnmarshalObjectOrString(t *testing.T) {
	var object RecipeAnswerSet
	if err := json.Unmarshal([]byte(`{"HOSTNAME":"web","YB_CPU_CORES":2}`), &object); err != nil {
		t.Fatal(err)
	}
	if object["HOSTNAME"] != "web" || object["YB_CPU_CORES"] != float64(2) {
		t.Fatalf("object = %#v", object)
	}

	var encoded RecipeAnswerSet
	if err := json.Unmarshal([]byte(`"{\"HOSTNAME\":\"web\"}"`), &encoded); err != nil {
		t.Fatal(err)
	}
	if encoded["HOSTNAME"] != "web" {
		t.Fatalf("string = %#v", encoded)
	}

	var empty RecipeAnswerSet
	if err := json.Unmarshal([]byte(`null`), &empty); err != nil {
		t.Fatal(err)
	}
	if empty != nil {
		t.Fatalf("null = %#v", empty)
	}
}

func TestRecipeQuestionJSON(t *testing.T) {
	payload := []byte(`{
		"$key": 4,
		"name": "YB_IP_ADDR_TYPE",
		"type": "list",
		"min": "",
		"max": "0",
		"default": null,
		"list": "{\"dhcp\":\"DHCP\",\"static\":\"Static\"}",
		"required": true
	}`)
	var question RecipeQuestion
	if err := json.Unmarshal(payload, &question); err != nil {
		t.Fatal(err)
	}
	if question.Min.Set {
		t.Fatal("empty min was set")
	}
	if !question.Max.Set || question.Max.N != 0 {
		t.Fatalf("max = %+v", question.Max)
	}
	if question.Choices["dhcp"] != "DHCP" {
		t.Fatalf("choices = %#v", question.Choices)
	}
	if recipeDefaultEmpty(question.Default) != true {
		t.Fatal("null default was not empty")
	}

	var recipe VMRecipe
	if err := json.Unmarshal([]byte(`{"$key":"abc","id":"abc","name":"Ubuntu","vm":null,"downloaded":false}`), &recipe); err != nil {
		t.Fatal(err)
	}
	if recipe.VM != nil || recipe.Key != "abc" {
		t.Fatalf("recipe = %+v", recipe)
	}
}

func TestParseRecipePreview(t *testing.T) {
	report := []byte(`{"err":"Simulation complete","response":{"cloudinit_files":[{"name":"user-data","contents":"#cloud-config"}],"logs":["rendered"],"answers":{"YB_VM_KEY":"99"}}}`)
	preview, err := parseRecipePreview(405, report, "preview")
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.CloudInitFiles) != 1 || preview.CloudInitFiles[0].Name != "user-data" {
		t.Fatalf("files = %#v", preview.CloudInitFiles)
	}
	if preview.Answers["YB_VM_KEY"] != "99" {
		t.Fatalf("answers = %#v", preview.Answers)
	}

	_, err = parseRecipePreview(200, []byte(`{"$key":1}`), "preview")
	if !IsRecipePreviewPersistedError(err) {
		t.Fatalf("persisted: %v", err)
	}

	failedBody := []byte(`{"err":"Simulation complete","response":{"cloudinit_files":[],"logs":["Error executing API command: no drive"],"answers":{}}}`)
	preview, err = parseRecipePreview(405, failedBody, "preview")
	if preview != nil || !IsRecipePreviewFailedError(err) {
		t.Fatalf("failed preview = %#v err = %v", preview, err)
	}
	var failed *RecipePreviewFailedError
	if !errors.As(err, &failed) || failed.Preview == nil || len(failed.Preview.Logs) != 1 {
		t.Fatalf("failed detail = %#v", err)
	}

	_, err = parseRecipePreview(405, []byte(`{"err":"Method not allowed"}`), "preview")
	if IsRecipePreviewPersistedError(err) || IsRecipePreviewFailedError(err) {
		t.Fatalf("other 405: %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 405 {
		t.Fatalf("api error = %#v", err)
	}
}
