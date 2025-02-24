package router

import (
	"strings"

	"realy.lol/realy/api"
)

type Protocol []api.Method

func Route(h api.H, p string) {
	acc := h.Request.Header.Get("Accept")
	c, ok := api.GetCapability(p)
	if !ok {
		log.W.F("unknown capability %s", p)
		return
	}
	if strings.HasPrefix(c.API(acc), h.URL.Path) {
		c.Handle(h)
	}
}
