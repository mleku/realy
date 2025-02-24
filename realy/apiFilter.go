package realy

import (
	acc "realy.lol/realy/accept"
	"realy.lol/realy/api"
)

type Filter struct{}

func (m *Filter) Handle(h api.H) {

}

func (m *Filter) API(accept string) (s string) {
	switch accept {
	case acc.NostrJSON:
		s = "todo"
	}
	return
}

func (m *Filter) Path() (s string) { return "/filter" }

func init() {
	api.RegisterCapability(&Filter{})
}
