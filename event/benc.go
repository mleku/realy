package event

import (
	"realy.lol/kind"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

func (ev *T) ToBenc() (benc *BencEvent) {
	benc = &BencEvent{
		Id:        ev.ID,
		Pubkey:    ev.PubKey,
		CreatedAt: ev.CreatedAt.I64(),
		Kind:      ev.Kind.ToI32(),
		Content:   ev.Content,
		Sig:       ev.Sig,
	}
	if ev.Tags.F() != nil {
		benc.Tags = make([]BencTag, len(ev.Tags.F()))
		for i, t := range ev.Tags.F() {
			benc.Tags[i] = BencTag{Tag: t.BS()}
		}
	}
	return
}

func (be *BencEvent) ToEvent() (ev *T) {
	ev = &T{
		ID:        be.Id,
		PubKey:    be.Pubkey,
		CreatedAt: timestamp.FromUnix(be.CreatedAt),
		Kind:      kind.New(be.Kind),
		Content:   be.Content,
		Sig:       be.Sig,
	}
	if len(be.Tags) != 0 {
		ev.Tags = tags.NewWithCap(len(be.Tags))
		for i := range be.Tags {
			ev.Tags.AppendSlice(be.Tags[i].Tag...)
		}
	}
	return
}
