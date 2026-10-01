package vergeos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// VMRecipeInstanceService deploys VMs from recipes and reads those deployments.
type VMRecipeInstanceService struct {
	client *Client
}

// List returns recipe instances.
// The default field set does not include answers.
func (s *VMRecipeInstanceService) List(ctx context.Context, opts ...ListOption) ([]VMRecipeInstance, error) {
	options := applyListOptions(opts)
	if options.Fields == "most" {
		options.Fields = vmRecipeInstanceListFields
	}

	var instances []VMRecipeInstance
	if err := s.client.get(ctx, "/vm_recipe_instances", options.toQueryParams(), &instances); err != nil {
		return nil, err
	}
	return instances, nil
}

// ListByRecipe returns instances deployed from the recipe with this hex key.
func (s *VMRecipeInstanceService) ListByRecipe(ctx context.Context, recipeID string, opts ...ListOption) ([]VMRecipeInstance, error) {
	if recipeID == "" {
		return nil, &ValidationError{Field: "recipe", Message: "recipe id is required"}
	}
	filter := WithFilter(fmt.Sprintf("recipe eq '%s'", escapeFilterValue(recipeID)))
	return s.List(ctx, append([]ListOption{filter}, opts...)...)
}

// Get returns one recipe instance, including its stored answers.
// Answers may contain a guest password.
func (s *VMRecipeInstanceService) Get(ctx context.Context, id int) (*VMRecipeInstance, error) {
	params := url.Values{}
	params.Set("fields", vmRecipeInstanceGetFields)
	var instance VMRecipeInstance
	endpoint := fmt.Sprintf("/vm_recipe_instances/%d", id)
	if err := s.client.get(ctx, endpoint, params, &instance); err != nil {
		if statusNotFound(err) {
			return nil, &NotFoundError{Resource: "VMRecipeInstance", ID: id}
		}
		return nil, err
	}
	return &instance, nil
}

// GetByName returns the recipe instance with this exact name.
func (s *VMRecipeInstanceService) GetByName(ctx context.Context, name string) (*VMRecipeInstance, error) {
	rows, err := s.List(ctx, WithFilter(fmt.Sprintf("name eq '%s'", escapeFilterValue(name))))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, &NotFoundError{Resource: "VMRecipeInstance", ID: name}
	}
	if err := requireUniqueName("VMRecipeInstance", name, rows, func(row VMRecipeInstance) any { return row.Key }); err != nil {
		return nil, err
	}
	if err := requireExactName("VMRecipeInstance", name, rows[0].Name, name); err != nil {
		return nil, err
	}
	got, err := s.Get(ctx, int(rows[0].Key))
	if err != nil {
		return nil, err
	}
	if err := requireExactName("VMRecipeInstance", name, got.Name, name); err != nil {
		return nil, err
	}
	return got, nil
}

// Deploy creates a VM from a recipe.
//
// Answers are checked against the recipe's questions before the request is
// sent. A bool answer has to be a boolean or one of true, yes, on, 1, false,
// no, off, 0, or the integers 0 and 1. Anything else is refused, because the
// platform would otherwise store it as false. A disksize answer is a byte
// count. Zero keeps the recipe default. Any other value under 1 MB is refused,
// because the platform would build the image's own disk and still report
// success. Network names are resolved to vnet keys. __new_internal__ is sent
// as itself.
//
// Deploy always creates. A dry run is Preview, which cannot return an instance.
func (s *VMRecipeInstanceService) Deploy(ctx context.Context, req *VMRecipeDeployRequest) (*VMRecipeInstance, error) {
	body, err := s.instanceBody(ctx, req, false)
	if err != nil {
		return nil, err
	}

	var resp apiResponse
	if err := s.client.post(ctx, "/vm_recipe_instances", body, &resp); err != nil {
		return nil, err
	}
	id, err := getKey(resp)
	if err != nil {
		return s.GetByName(ctx, req.Name)
	}
	return s.Get(ctx, id)
}

// Preview asks the platform to walk a deploy and returns the report.
//
// It sends the same checked answers as Deploy and sets the platform's
// simulation flag. It does not return a VMRecipeInstance, and it does not
// look one up by name. A successful HTTP status means the platform accepted
// a real create. That is RecipePreviewPersistedError: a VM with this name
// may exist, and this call will not hand it back as a preview.
//
// The platform's normal preview reply is HTTP 405 with err "Simulation complete"
// and a report of cloud-init files, logs, and resolved answers. Keys in that
// report are predictions. Log lines that record a failed step are
// RecipePreviewFailedError. The report on that error may contain guest credentials.
func (s *VMRecipeInstanceService) Preview(ctx context.Context, req *VMRecipeDeployRequest) (*VMRecipePreview, error) {
	body, err := s.instanceBody(ctx, req, true)
	if err != nil {
		return nil, err
	}
	status, respBody, err := s.client.postStatus(ctx, "/vm_recipe_instances", body)
	if err != nil {
		return nil, err
	}
	name := ""
	if req != nil {
		name = req.Name
	}
	return parseRecipePreview(status, respBody, name)
}

// instanceBody builds the POST document. simulate is true only for Preview.
// The public request has no field that can turn a deploy into a preview.
func (s *VMRecipeInstanceService) instanceBody(ctx context.Context, req *VMRecipeDeployRequest, simulate bool) (map[string]any, error) {
	if req == nil {
		return nil, &ValidationError{Message: "request is required"}
	}
	if req.Recipe == "" {
		return nil, &ValidationError{Field: "recipe", Message: "recipe is required"}
	}
	if req.Name == "" {
		return nil, &ValidationError{Field: "name", Message: "name is required"}
	}

	questions, err := s.client.VMRecipes.Questions(ctx, req.Recipe)
	if err != nil {
		return nil, err
	}
	var networks []Network
	if recipeAnswersNeedNetworks(questions, req.Answers) {
		networks, err = s.client.Networks.List(ctx, WithFields("$key,name"))
		if err != nil {
			return nil, err
		}
	}
	resolved, err := resolveRecipeAnswers(questions, req.Answers, networks)
	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"recipe": req.Recipe,
		"name":   req.Name,
	}
	if len(resolved) > 0 {
		body["answers"] = resolved
	}
	if req.AutoUpdate != nil {
		body["auto_update"] = *req.AutoUpdate
	}
	if simulate {
		body["simulate"] = true
	}
	return body, nil
}

func parseRecipePreview(status int, body []byte, name string) (*VMRecipePreview, error) {
	if status >= 200 && status < 300 {
		return nil, &RecipePreviewPersistedError{Name: name}
	}

	var envelope struct {
		Err      string          `json:"err"`
		Response json.RawMessage `json:"response"`
	}
	unmarshalErr := json.Unmarshal(body, &envelope)
	if unmarshalErr != nil || envelope.Err != recipePreviewCompleteMarker || !previewReportShape(envelope.Response) {
		message := strings.TrimSpace(string(body))
		if unmarshalErr == nil && envelope.Err != "" {
			message = envelope.Err
		}
		return nil, apiStatusError(status, "/vm_recipe_instances", message)
	}

	var preview VMRecipePreview
	if err := json.Unmarshal(envelope.Response, &preview); err != nil {
		return nil, fmt.Errorf("vergeos: failed to decode recipe preview: %w", err)
	}
	if lines := previewLogErrors(preview.Logs); len(lines) > 0 {
		return nil, &RecipePreviewFailedError{Lines: lines, Preview: &preview}
	}
	return &preview, nil
}

func previewReportShape(raw json.RawMessage) bool {
	if len(strings.TrimSpace(string(raw))) == 0 {
		return false
	}
	var obj map[string]json.RawMessage
	if json.Unmarshal(raw, &obj) != nil {
		return false
	}
	files, filesOK := obj["cloudinit_files"]
	logs, logsOK := obj["logs"]
	answers, answersOK := obj["answers"]
	if !filesOK || !logsOK || !answersOK {
		return false
	}
	if string(files) == "null" || string(logs) == "null" || string(answers) == "null" {
		return false
	}
	var fileList []json.RawMessage
	var logList []json.RawMessage
	var answerObj map[string]json.RawMessage
	if json.Unmarshal(files, &fileList) != nil || json.Unmarshal(logs, &logList) != nil || json.Unmarshal(answers, &answerObj) != nil {
		return false
	}
	return true
}
