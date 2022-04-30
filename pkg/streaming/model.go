package streaming

type ExceptionDto struct {
	Message    string `json:"message"`
	StackTrace string `json:"stacktrace"`
}

type Notification struct {
	Time      int64        `json:"time"`
	Severity  string       `json:"severity"`
	Source    string       `json:"source"`
	Key       string       `json:"key"`
	Value     []byte       `json:"value"`
	Message   string       `json:"message"`
	Exception ExceptionDto `json:"exception,omitempty"`
}
