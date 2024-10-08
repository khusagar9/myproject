package models

type ApiResponse struct {
	Offset      int           `json:"offset"`
	Limit       int           `json:"limit"`
	Amount      int           `json:"amount"`
	Total       int           `json:"total"`
	CeInstances []CeInstances `json:"ceInstances"`
}

type CeInstances struct {
	ID       string   `json:"id"`
	Geometry Geometry `json:"geometry"`
	AreaPath string   `json:"areaPath"`
	Data     Data     `json:"data"`
}

type Geometry struct {
	Coordinates interface{} `json:"coordinates"`
	Type        string      `json:"type"`
}

type Data struct {
	ActivationEnd   ActivationTime `json:"activationEnd"`
	ActivationStart ActivationTime `json:"activationStart"`
}

type ActivationTime struct {
	TimeString  string `json:"timeString"`
	TimestampMs int64  `json:"timestampMs"`
}
