package pointers

import (
	"realy.lol/eventid"
	"realy.lol/kind"
)

type Profile struct {
	PublicKey []byte   `json:"pubkey"`
	Relays    [][]byte `json:"relays,omitempty"`
}

type Event struct {
	ID     *eventid.T `json:"id"`
	Relays [][]byte   `json:"relays,omitempty"`
	Author []byte     `json:"author,omitempty"`
	Kind   *kind.T    `json:"kind,omitempty"`
}

type Entity struct {
	PublicKey  []byte   `json:"pubkey"`
	Kind       *kind.T  `json:"kind,omitempty"`
	Identifier []byte   `json:"identifier,omitempty"`
	Relays     [][]byte `json:"relays,omitempty"`
}
