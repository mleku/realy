package nip86

type IDReason struct {
	ID     st `json:"id"`
	Reason st `json:"reason"`
}

type PubKeyReason struct {
	PubKey st `json:"pubkey"`
	Reason st `json:"reason"`
}

type IPReason struct {
	IP     st `json:"ip"`
	Reason st `json:"reason"`
}
