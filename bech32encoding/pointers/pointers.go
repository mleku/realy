package pointers

import (
	"realy.lol/eventid"
	"realy.lol/kind"
)

type Profile struct {
	PublicKey by   `json:"pubkey"`
	Relays    []by `json:"relays,omitempty"`
}

type Event struct {
	ID     *eventid.T `json:"id"`
	Relays []by       `json:"relays,omitempty"`
	Author by         `json:"author,omitempty"`
	Kind   *kind.T    `json:"kind,omitempty"`
}

type Entity struct {
	PublicKey  by      `json:"pubkey"`
	Kind       *kind.T `json:"kind,omitempty"`
	Identifier by      `json:"identifier,omitempty"`
	Relays     []by    `json:"relays,omitempty"`
}
