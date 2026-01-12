package common

// Deprecated
type TelemetryConf struct {
	LogLevel      string `json:"logLevel"` //日志级别，从0~7，0代表全部输出，7代表关闭输出。all,trace,debug,info,warn,error,fatal,off
	TraceUrl      string `json:"traceUrl"`
	LogUrl        string `json:"logUrl"`
	ServerName    string `json:"serverName"`
	ServerVersion string `json:"serverVersion"`
	TraceEnabled  bool   `json:"traceEnabled,string"`
	// AuditUrl 审计日志的上报地址，例如 https://anyrobot.example.org/api/feed_ingester/v1/jobs/12d627fb523d7362/events
	AuditUrl string `json:"auditUrl"`
	//AuditEnabled  是否启用审计日志
	AuditEnabled bool `json:"auditEnabled,string"`
}

const (
	ALL   = "all"
	TRACE = "trace"
	DEBUG = "debug"
	INFO  = "info"
	WARN  = "warn"
	ERROR = "error"
	FATAL = "fatal"
	OFF   = "off"
)
