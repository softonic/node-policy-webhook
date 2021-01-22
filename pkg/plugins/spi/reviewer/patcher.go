package reviewer

// Patcher generate patch given the context
type Patcher interface {
	Create() ([]PatchOperation, error)
}

// PatchOperation one patch step
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}
