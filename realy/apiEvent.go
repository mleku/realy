package realy

import (
	acc "realy.lol/realy/accept"
	"realy.lol/realy/api"
)

type Event struct{}

func (m *Event) Handle(h api.H) {

}

func (m *Event) API(accept string) (s string) {
	switch accept {
	case acc.NostrJSON:
		s = "todo"
	}
	return
}

func (m *Event) Path() (s string) { return "/event" }

func init() {
	api.RegisterCapability(&Event{})
}
