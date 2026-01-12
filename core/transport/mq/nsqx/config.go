package nsqx

type Config struct {
	Host        string `json:"host"`
	HttpHost    string `json:"httpHost"`
	LookupdHost string `json:"lookupdHost"`
	Channel     string `json:"channel"`
}
