package realy

import (
	acc "realy.lol/realy/accept"
	"realy.lol/realy/api"
)

type API struct{}

func (m *API) Handle(h api.H) {

}

func (m *API) API(accept string) (s string) {
	switch accept {
	case acc.NostrJSON:
		s = "todo"
	}
	return
}

func (m *API) Path() (s string) { return "/api" }

func init() {
	api.RegisterCapability(&API{})
}
