package vergeos

// Member represents a group membership in VergeOS.
type Member struct {
	// ID is the unique identifier for the membership.
	ID FlexInt `json:"$key,omitempty"`
	// Group is the parent group ID.
	Group int `json:"parent_group,omitempty"`
	// Member is the member value (user or group reference).
	Member string `json:"member,omitempty"`
}

// MemberCreateRequest is the request body for creating a membership.
type MemberCreateRequest struct {
	// Group is the parent group ID (required).
	Group int `json:"parent_group"`
	// Member is the member value (required).
	Member string `json:"member"`
}

// MemberUpdateRequest is the request body for updating a membership.
type MemberUpdateRequest struct {
	// Group is the parent group ID.
	Group *int `json:"parent_group,omitempty"`
	// Member is the member value.
	Member *string `json:"member,omitempty"`
}

// memberListFields are the fields to request when listing members.
const memberListFields = "$key,parent_group,member"
