package data

type DataRequest struct {
	Key   string `json:"key"`
	Data  string `json:"data"`
	Owner string `json:"owner"`
}

type JoinRequest struct {
	Key  string `json:"key"`
	Host string `json:"host"`
}
