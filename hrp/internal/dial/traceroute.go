package dial

type TraceRouteOptions struct {
	MaxTTL    int
	Queries   int
	SaveTests bool
}

type TraceRouteResult struct {
	IP      string                 `json:"ip"`
	Details []TraceRouteResultNode `json:"details"`
	Suc     bool                   `json:"suc"`
	ErrMsg  string                 `json:"errMsg"`
}

type TraceRouteResultNode struct {
	Id   int    `json:"id"`
	Ip   string `json:"ip"`
	Time string `json:"time"`
}
