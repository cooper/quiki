package wikifier

import (
	"encoding/json"
)

// MarshalJSON returns a JSON representation of the page.
// It includes the PageInfo, HTML, and CSS.
func (p *Page) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		HTML HTML   `json:"html"`
		CSS  string `json:"css"`
		PageInfo
	}{
		HTML:     p.HTML(),
		CSS:      p.CSS(),
		PageInfo: p.Info(),
	})
}
