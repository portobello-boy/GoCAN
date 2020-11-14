package data

type DataRequest struct {
	Key   string `json:"key"`
	Data  string `json:"data"`
	Owner string `json:"owner"`
}

type DataResponse struct {
	Key    string    `json:"key"`
	Data   string    `json:"data"`
	Coords []float64 `json:"coords"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type JoinRequest struct {
	Key  string `json:"key"`
	Host string `json:"host"`
}
