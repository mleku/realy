package realy

import (
	acc "realy.lol/realy/accept"
	"realy.lol/realy/api"
)

type Fulltext struct{}

func (m *Fulltext) Handle(h api.H) {

}

func (m *Fulltext) API(accept string) (s string) {
	switch accept {
	case acc.NostrJSON:
		s = "todo"
	}
	return
}

func (m *Fulltext) Path() (s string) { return "/fulltext" }

func init() {
	api.RegisterCapability(&Fulltext{})
}
