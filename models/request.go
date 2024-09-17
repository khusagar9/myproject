package models

// RequestBody represents the overall structure of the request.
type RequestBody struct {
	CeInstance CeInstance `json:"ceInstance"`
	Offset     int        `json:"offset"`
	Limit      int        `json:"limit"`
	Order      string     `json:"order"`
	WithTotal  bool       `json:"withTotal"`
}

// CeInstance contains the nested structure for ceInstance field.
type CeInstance struct {
	TemplateID TemplateID `json:"templateId"`
}

// TemplateID contains the equals field, which is an array of strings.
type TemplateID struct {
	Equals []string `json:"equals"`
}
