package nip86

type IDReason struct {
	ID     S `json:"id"`
	Reason S `json:"reason"`
}

type PubKeyReason struct {
	PubKey S `json:"pubkey"`
	Reason S `json:"reason"`
}

type IPReason struct {
	IP     S `json:"ip"`
	Reason S `json:"reason"`
}
