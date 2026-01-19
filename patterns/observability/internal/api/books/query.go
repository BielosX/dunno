package books

type SearchRequest struct {
	Size        int32            `json:"size"`
	Sort        []map[string]any `json:"sort"`
	Query       QueryPayload     `json:"query"`
	SearchAfter []string         `json:"search_after,omitempty"`
}

type QueryPayload struct {
	Bool BoolPayload `json:"bool"`
}

type BoolPayload struct {
	Filter []map[string]any `json:"filter"`
}

func (p *BoolPayload) appendFilter(v map[string]any) {
	p.Filter = append(p.Filter, v)
}

func (p *BoolPayload) appendTermFilter(v map[string]any) {
	p.Filter = append(p.Filter, map[string]any{
		"term": v,
	})
}

func (p *BoolPayload) appendPrefixFilter(v map[string]any) {
	p.Filter = append(p.Filter, map[string]any{
		"prefix": v,
	})
}
