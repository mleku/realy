package event

import (
	"realy.lol/ec/schnorr"
	"realy.lol/kind"
	"realy.lol/sha256"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

// ToProto converts from the event.T to the protobuf event format.
func (ev *T) ToProto() (proto *Event) {
	proto = &Event{
		Id:        ev.ID,
		Pubkey:    ev.PubKey,
		CreatedAt: ev.CreatedAt.I64(),
		Kind:      int32(ev.Kind.ToInt()),
		Content:   ev.Content,
		Sig:       ev.Sig,
	}
	if ev.Tags.Len() > 0 {
		//log.I.S(ev.Tags.F())
		proto.Tags = make([]*Tag, ev.Tags.Len())
		for i, v := range ev.Tags.F() {
			proto.Tags[i] = &Tag{Fields: v.F()}
		}
	}
	return
}

// ToEvent converts from the protobuf event format to the event.T.
func (x *Event) ToEvent() (ev *T) {
	if x == nil || x.Id == nil {
		return &T{}
	}
	ev = &T{
		ID:        x.Id[:sha256.Size],
		PubKey:    x.Pubkey[:schnorr.PubKeyBytesLen],
		CreatedAt: timestamp.FromUnix(x.CreatedAt),
		Kind:      kind.New(x.Kind),
		Content:   x.Content,
		Sig:       x.Sig[:schnorr.SignatureSize],
	}
	if x.Tags != nil {
		ev.Tags = tags.NewWithCap(len(x.Tags))
		for i := range x.Tags {
			ev.Tags.AppendSlice(x.Tags[i].Fields...)
		}
	}
	return
}
