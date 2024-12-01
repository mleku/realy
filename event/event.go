package event

import (
	"lukechampine.com/frand"

	"realy.lol/eventid"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/sha256"
	"realy.lol/signer"
	"realy.lol/tags"
	"realy.lol/text"
	"realy.lol/timestamp"
)

// T is the primary datatype of nostr. This is the form of the structure that
// defines its JSON string based format.
type T struct {
	// ID is the SHA256 hash of the canonical encoding of the event in binary format
	ID by
	// PubKey is the public key of the event creator in binary format
	PubKey by
	// CreatedAt is the UNIX timestamp of the event according to the event
	// creator (never trust a timestamp!)
	CreatedAt *timestamp.T
	// Kind is the nostr protocol code for the type of event. See kind.T
	Kind *kind.T
	// Tags are a list of tags, which are a list of strings usually structured
	// as a 3 layer scheme indicating specific features of an event.
	Tags *tags.T
	// Content is an arbitrary string that can contain anything, but usually
	// conforming to a specification relating to the Kind and the Tags.
	Content by
	// Sig is the signature on the ID hash that validates as coming from the
	// Pubkey in binary format.
	Sig by
}

// Ts is an array of T that sorts in reverse chronological order.
type Ts []*T

func (ev Ts) Len() no         { return len(ev) }
func (ev Ts) Less(i, j no) bo { return *ev[i].CreatedAt > *ev[j].CreatedAt }
func (ev Ts) Swap(i, j no)    { ev[i], ev[j] = ev[j], ev[i] }

type C chan *T

func New() (ev *T) { return &T{} }

func (ev *T) Serialize() (b by) {
	b, _ = ev.MarshalJSON(nil)
	return
}

// func (ev *T) String() (r S) { return S(ev.Serialize()) }

func (ev *T) ToCanonical() (b by) {
	b = append(b, "[0,\""...)
	b = hex.EncAppend(b, ev.PubKey)
	b = append(b, "\","...)
	var err er
	if b, err = ev.CreatedAt.MarshalJSON(b); chk.E(err) {
		return
	}
	b = append(b, ',')
	if b, err = ev.Kind.MarshalJSON(b); chk.E(err) {
		return
	}
	b = append(b, ',')
	if b, err = ev.Tags.MarshalJSON(b); chk.E(err) {
		panic(err)
	}
	b = append(b, ',')
	b = text.AppendQuote(b, ev.Content, text.NostrEscape)
	b = append(b, ']')
	return
}

// stringy functions for retarded other libraries

func (ev *T) IDString() (s st)          { return hex.Enc(ev.ID) }
func (ev *T) EventID() (eid *eventid.T) { return eventid.NewWith(ev.ID) }
func (ev *T) PubKeyString() (s st)      { return hex.Enc(ev.PubKey) }
func (ev *T) SigString() (s st)         { return hex.Enc(ev.Sig) }
func (ev *T) TagStrings() (s [][]st) {
	return ev.Tags.ToStringSlice()
}
func (ev *T) ContentString() (s st) { return st(ev.Content) }

func Hash(in by) (out by) {
	h := sha256.Sum256(in)
	return h[:]
}

// GetIDBytes returns the raw SHA256 hash of the canonical form of an T.
func (ev *T) GetIDBytes() by { return Hash(ev.ToCanonical()) }

func GenerateRandomTextNoteEvent(sign signer.I, maxSize no) (ev *T,
	err er) {

	l := frand.Intn(maxSize * 6 / 8) // account for base64 expansion
	ev = &T{
		PubKey:    sign.Pub(),
		Kind:      kind.TextNote,
		CreatedAt: timestamp.Now(),
		Content:   text.NostrEscape(nil, frand.Bytes(l)),
		Tags:      tags.New(),
	}
	if err = ev.Sign(sign); chk.E(err) {
		return
	}
	return
}
