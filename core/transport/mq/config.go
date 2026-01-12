package mq

type MQConf struct {
	ConnConfs   []*MQConnConf `json:"connConfs"`
	Channel     string        `json:"channel"`
	ClientID    string        `json:"clientId"`
	SendBufSize int           `json:"sendBufSize"`
	RecvBufSize int           `json:"recvBufSize"`
}

func (c MQConf) GetMQConnConfByMQType(mqType string) *MQConnConf {
	for i := range c.ConnConfs {
		if c.ConnConfs[i].MQType == mqType {
			return c.ConnConfs[i]
		}
	}
	return nil
}

type MQConnConf struct {
	MQType      string `json:"mqType"`
	Addr        string `json:"host"`
	HttpHost    string `json:"httpHost"`
	LookupdAddr string `json:"lookupdHost"`
	MQAuthConf  `json:"auth"`
}

type MQAuthConf struct {
	Mechanism string `json:"mechanism"`
	User      string `json:"username"`
	Password  string `json:"password"`
	Version   string `json:"version"`
	ClientID  string `json:"client_id"`
	GroupID   string `json:"group_id"`
}
