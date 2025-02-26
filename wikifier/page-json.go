package wikifier

import (
	"encoding/json"
)

func (p *Page) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Title string `json:"title"`
		HTML  HTML   `json:"html"`
		CSS   string `json:"css"`
		PageInfo
	}{
		HTML:     p.HTML(),
		CSS:      p.CSS(),
		PageInfo: p.Info(),
	})
}
