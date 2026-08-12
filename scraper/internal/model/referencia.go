package model

// Referencia apunta al mismo punto tal como lo reporta otra fuente.
type Referencia struct {
	Fuente string `json:"fuente"`
	ID     string `json:"id"`
	URL    string `json:"url,omitempty"`
}
