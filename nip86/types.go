package nip86

type Request struct {
	Method S     `json:"method"`
	Params []any `json:"params"`
}

type Response struct {
	Result any `json:"result,omitempty"`
	Error  S   `json:"error,omitempty"`
}
