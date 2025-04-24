package openapi

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/context"
	"realy.mleku.dev/eventidserial"
	"realy.mleku.dev/realy/helpers"
)

type EventIdsBySerialInput struct {
	Start uint64 `path:"start" doc:"fetch events by their database serial beginning at this number" required:"true"`
	Count int    `path:"count" doc:"maximum number of events to return (max 1000)" required:"true"`
}

type EventIdsBySerialOutputBody struct {
	Events []eventidserial.E `doc:"event serials andIDs"`
}

type EventIdsBySerialOutput struct {
	Body []eventidserial.E `doc:"event serials andIDs"`
}

// RegisterEventIdsBySerial is a query that allows fetching of events based on their internal serial
// number, enabling simple synchronisation of the database of events in the order they were
// stored.
//
// If the event doesn't exist, the next largest one that does is the start and the count
// parameter says how many.
//
// Results are a json object with keys as stringified versions of the serial numbers, in
// ascending order.
//
// Serials are guaranteed to be stable and either refer to the same event or not exist anymore
// if the event was deleted, and start at zero and count up monotonically, and atomically.
//
// Requests are limited to 512 events per request, and the next number after the last result
// will refer to the next event in the sequence of serial numbers of events that exist.
func (x *Operations) RegisterEventIdsBySerial(api huma.API) {
	name := "EventIdsBySerial"
	description := "Fetch event IDs by their database serial, to get the events use /event"
	path := x.path + "/eventidsbyserial/{start}/{count}"
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
	}, func(ctx context.T, input *EventIdsBySerialInput) (output *EventIdsBySerialOutput, err error) {
		if !x.Server.Configured() {
			err = huma.Error503ServiceUnavailable("server is not configured")
			return
		}
		var out []eventidserial.E
		if out, err = x.Server.Storage().EventIdsBySerial(input.Start, input.Count); chk.E(err) {
			return
		}
		output = new(EventIdsBySerialOutput)
		for _, e := range out {
			output.Body = append(output.Body, e)
		}
		return
	})
}
