// Package entity defines Ref, the plain-data reference to "a thing the user is
// looking at" that tuigoff components pass around: the row under a grid's
// cursor, an entry pinned in a sidebar. It carries no behaviour beyond
// equality and knows nothing about sessions, LLMs or products; higher layers
// (for example strongo/aichat's session.EntityRef) alias or convert to it.
package entity

// Ref is a stable reference to a product entity.
type Ref struct {
	// Type names the kind of entity, e.g. "happening" or "recordset".
	Type string `json:"type" yaml:"type"`
	// Keys identify the entity within its Type, e.g. {"spaceID": "...", "id": "..."}.
	Keys map[string]string `json:"keys" yaml:"keys"`
	// Title is a display hint (a sidebar row, a prompt). It may be stale and
	// is ignored by Same.
	Title string `json:"title,omitempty" yaml:"title,omitempty"`
}

// Same reports whether two refs point at the same entity (Title ignored).
func (r Ref) Same(o Ref) bool {
	if r.Type != o.Type || len(r.Keys) != len(o.Keys) {
		return false
	}
	for k, v := range r.Keys {
		if ov, ok := o.Keys[k]; !ok || ov != v {
			return false
		}
	}
	return true
}
