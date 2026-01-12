package common

type ServerConf struct {
	HttpConf `json:"http"`
	SwagConf `json:"doc"`
}
type HttpConf struct {
	Host      string `json:"host"`
	InnerHost string `json:"innerHost"`
}

type SwagConf struct {
	Host    string `json:"host"`
	Version string `json:"version"`
}
