package event

import (
	"reflect"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/codec"
	"realy.mleku.dev/errorf"
	"realy.mleku.dev/hex"
	"realy.mleku.dev/json"
	"realy.mleku.dev/kind"
	"realy.mleku.dev/log"
	"realy.mleku.dev/tags"
	"realy.mleku.dev/text"
	"realy.mleku.dev/timestamp"
)

// ToCanonical converts the event to the canonical encoding used to derive the
// event Id.
func (ev *T) ToCanonical(dst []byte) (b []byte) {
	b = dst
	b = append(b, "[0,\""...)
	b = hex.EncAppend(b, ev.Pubkey)
	b = append(b, "\","...)
	b = ev.CreatedAt.Marshal(b)
	b = append(b, ',')
	b = ev.Kind.Marshal(b)
	b = append(b, ',')
	b = ev.Tags.Marshal(b)
	b = append(b, ',')
	b = text.AppendQuote(b, ev.Content, text.NostrEscape)
	b = append(b, ']')
	return
}

// GetIDBytes returns the raw SHA256 hash of the canonical form of an event.T.
func (ev *T) GetIDBytes() []byte { return Hash(ev.ToCanonical(nil)) }

// NewCanonical builds a new canonical encoder.
func NewCanonical() (a *json.Array) {
	a = &json.Array{
		V: []codec.JSON{
			&json.Unsigned{}, // 0
			&json.Hex{},      // pubkey
			&timestamp.T{},   // created_at
			&kind.T{},        // kind
			&tags.T{},        // tags
			&json.String{},   // content
		},
	}
	return
}

// this is an absolute minimum length canonical encoded event
var minimal = len(`[0,"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",1733739427,0,[],""]`)

// FromCanonical reverses the process of creating the canonical encoding, note
// that the signature is missing in this form. Allocate an event.T before
// calling this.
func (ev *T) FromCanonical(b []byte) (rem []byte, err error) {
	rem = b
	id := Hash(rem)
	c := NewCanonical()
	if rem, err = c.Unmarshal(rem); chk.E(err) {
		log.I.F("%s", b)
		return
	}
	// unwrap the array
	x := (*c).V
	if v, ok := x[0].(*json.Unsigned); !ok {
		err = errorf.E("did not decode expected type in first field of canonical event %v %v",
			reflect.TypeOf(x[0]), x[0])
		return
	} else {
		if v.V != 0 {
			err = errorf.E("unexpected value %d in first field of canonical event, expect 0",
				v.V)
			return
		}
	}
	// create the event, use the Id hash to populate the Id
	ev.Id = id
	// unwrap the pubkey
	if v, ok := x[1].(*json.Hex); !ok {
		err = errorf.E("failed to decode pubkey from canonical form of event %s", b)
		return
	} else {
		ev.Pubkey = v.V
	}
	// populate the timestamp field
	if v, ok := x[2].(*timestamp.T); !ok {
		err = errorf.E("did not decode expected type in third (created_at) field of canonical event %v %v",
			reflect.TypeOf(x[0]), x[0])
	} else {
		ev.CreatedAt = v
	}
	// populate the kind field
	if v, ok := x[3].(*kind.T); !ok {
		err = errorf.E("did not decode expected type in fourth (kind) field of canonical event %v %v",
			reflect.TypeOf(x[0]), x[0])
	} else {
		ev.Kind = v
	}
	// populate the tags field
	if v, ok := x[4].(*tags.T); !ok {
		err = errorf.E("did not decode expected type in fifth (tags) field of canonical event %v %v",
			reflect.TypeOf(x[0]), x[0])
	} else {
		ev.Tags = v
	}
	// populate the content field
	if v, ok := x[5].(*json.String); !ok {
		err = errorf.E("did not decode expected type in sixth (content) field of canonical event %v %v",
			reflect.TypeOf(x[0]), x[0])
	} else {
		ev.Content = v.V
	}
	return
}
