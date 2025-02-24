package realy

import (
	acc "realy.lol/realy/accept"
	"realy.lol/realy/api"
)

type Subscribe struct{}

func (m *Subscribe) Handle(h api.H) {

}

func (m *Subscribe) API(accept string) (s string) {
	switch accept {
	case acc.NostrJSON:
		s = "todo"
	}
	return
}

func (m *Subscribe) Path() (s string) { return "/subscribe" }

func init() {
	api.RegisterCapability(&Subscribe{})
}
