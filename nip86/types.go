package nip86

type Request struct {
	Method st    `json:"method"`
	Params []any `json:"params"`
}

type Response struct {
	Result any `json:"result,omitempty"`
	Error  st  `json:"error,omitempty"`
}
