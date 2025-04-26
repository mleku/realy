package openapi

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"realy.lol/chk"
	"realy.lol/context"
	"realy.lol/log"
	"realy.lol/realy/helpers"
)

type EventCountOutput struct{ Body uint64 }

func (x *Operations) RegisterEventCount(api huma.API) {
	name := "EventCount"
	description := "Report the number of events in the database"
	path := x.path + "/eventcount"
	scopes := []string{"user", "read"}
	method := http.MethodGet
	huma.Register(api, huma.Operation{
		OperationID: name,
		Summary:     name,
		Path:        path,
		Method:      method,
		Tags:        []string{"events"},
		Description: helpers.GenerateDescription(description, scopes),
		Security:    []map[string][]string{{"auth": scopes}},
	}, func(ctx context.T, input *struct{}) (output *EventCountOutput, err error) {
		var c uint64
		now := time.Now()
		if c, err = x.Server.Storage().EventCount(); chk.E(err) {
			err = huma.Error500InternalServerError(err.Error())
			return
		}
		log.I.F("event count %d in %v", c, time.Now().Sub(now))
		output = &EventCountOutput{Body: c}
		return
	})
}
