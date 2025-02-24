package realy

import (
	acc "realy.lol/realy/accept"
	"realy.lol/realy/api"
)

type Relay struct{}

func (m *Relay) Handle(h api.H) {

}

func (m *Relay) API(accept string) (s string) {
	switch accept {
	case acc.NostrJSON:
		s = "todo"
	}
	return
}

func (m *Relay) Path() (s string) { return "/relay" }

func init() {
	api.RegisterCapability(&Relay{})
}
