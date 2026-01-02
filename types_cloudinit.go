package vergeos

// CloudInitFile represents a cloud-init file in VergeOS.
type CloudInitFile struct {
	// ID is the unique identifier for the file.
	ID FlexInt `json:"$key,omitempty"`
	// Owner is the owner of the file.
	Owner string `json:"owner,omitempty"`
	// Name is the file name.
	Name string `json:"name"`
	// FileSize is the file size in bytes.
	FileSize int64 `json:"filesize,omitempty"`
	// Contents is the file contents.
	Contents string `json:"contents,omitempty"`
	// ContainsVariables indicates whether the file contains variables.
	ContainsVariables bool `json:"containsVariables,omitempty"`
}

// CloudInitFileCreateRequest is the request body for creating a cloud-init file.
type CloudInitFileCreateRequest struct {
	// Name is the file name (required).
	Name string `json:"name"`
	// Contents is the file contents.
	Contents string `json:"contents,omitempty"`
	// ContainsVariables indicates whether the file contains variables.
	ContainsVariables *bool `json:"contains_variables,omitempty"`
}

// CloudInitFileUpdateRequest is the request body for updating a cloud-init file.
type CloudInitFileUpdateRequest struct {
	// Name is the file name.
	Name *string `json:"name,omitempty"`
	// Contents is the file contents.
	Contents *string `json:"contents,omitempty"`
	// ContainsVariables indicates whether the file contains variables.
	ContainsVariables *bool `json:"contains_variables,omitempty"`
	// FileSize is the file size.
	FileSize *int64 `json:"filesize,omitempty"`
}

// cloudInitListFields are the fields to request when listing cloud-init files.
const cloudInitListFields = "$key,name,filesize,contents,containsVariables"
