package data

type PointResponse struct {
	Coords []float64 `json:"coords"`
}

type RangeResponse struct {
	P1 PointResponse `json:"p1"`
	P2 PointResponse `json:"p2"`
}

type DataRequest struct {
	Key   string `json:"key"`
	Data  string `json:"data"`
	Owner string `json:"owner"`
}

type DataResponse struct {
	Key     string    `json:"key"`
	Data    string    `json:"data"`
	Coords  []float64 `json:"coords"`
	Message string    `json:"message"`
}

type DebugResponse struct {
	Dimension  int               `json:"dimension"`
	Redundancy int               `json:"redundancy"`
	Range      RangeResponse     `json:"range"`
	Data       map[string]string `json:"data"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type JoinRequest struct {
	Key  string `json:"key"`
	Host string `json:"host"`
}
