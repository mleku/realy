package realy

import (
	acc "realy.lol/realy/accept"
	"realy.lol/realy/api"
)

type Events struct{}

func (m *Events) Handle(h api.H) {

}

func (m *Events) API(accept string) (s string) {
	switch accept {
	case acc.NostrJSON:
		s = "todo"
	}
	return
}

func (m *Events) Path() (s string) { return "/events" }

func init() {
	api.RegisterCapability(&Events{})
}
