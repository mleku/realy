package event

import (
	"lukechampine.com/frand"

	"realy.lol/eventid"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/sha256"
	"realy.lol/signer"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/text"
	"realy.lol/timestamp"
)

// T is the primary datatype of nostr. This is the form of the structure that
// defines its JSON string based format.
type T struct {
	// ID is the SHA256 hash of the canonical encoding of the event in binary format
	ID []byte
	// PubKey is the public key of the event creator in binary format
	PubKey []byte
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
	Content []byte
	// Sig is the signature on the ID hash that validates as coming from the
	// Pubkey in binary format.
	Sig []byte
}

// Ts is an array of T that sorts in reverse chronological order.
type Ts []*T

func (ev Ts) Len() int           { return len(ev) }
func (ev Ts) Less(i, j int) bool { return ev[i].CreatedAt.I64() > ev[j].CreatedAt.I64() }
func (ev Ts) Swap(i, j int)      { ev[i], ev[j] = ev[j], ev[i] }

type C chan *T

func New() (ev *T) { return &T{} }

func (ev *T) Serialize() (b []byte) { return ev.Marshal(nil) }

func (ev *T) SerializeIndented() (b []byte) { return ev.marshalWithWhitespace(nil, true) }

func (ev *T) EventID() (eid *eventid.T) { return eventid.NewWith(ev.ID) }

// stringy/numbery functions for retarded other libraries

func (ev *T) IDString() (s string)       { return hex.Enc(ev.ID) }
func (ev *T) CreatedAtInt64() (i int64)  { return ev.CreatedAt.I64() }
func (ev *T) KindInt32() (i int32)       { return int32(ev.Kind.K) }
func (ev *T) PubKeyString() (s string)   { return hex.Enc(ev.PubKey) }
func (ev *T) SigString() (s string)      { return hex.Enc(ev.Sig) }
func (ev *T) TagStrings() (s [][]string) { return ev.Tags.ToStringSlice() }
func (ev *T) ContentString() (s string)  { return string(ev.Content) }

type J struct {
	Id        string     `json:"id"`
	Pubkey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int32      `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

func (ev *T) ToEventJ() (j *J) {
	j = &J{}
	j.Id = ev.IDString()
	j.Pubkey = ev.PubKeyString()
	j.CreatedAt = ev.CreatedAt.I64()
	j.Kind = ev.KindInt32()
	j.Content = ev.ContentString()
	j.Tags = ev.Tags.ToStringSlice()
	j.Sig = ev.SigString()
	return
}

func (ev *T) IDFromString(s string) (err error) {
	ev.ID, err = hex.Dec(s)
	return
}

func (ev *T) CreatedAtFromInt64(i int64) {
	ev.CreatedAt = timestamp.FromUnix(i)
	return
}

func (ev *T) KindFromInt32(i int32) {
	ev.Kind = &kind.T{}
	ev.Kind.K = uint16(i)
	return
}

func (ev *T) PubKeyFromString(s string) (err error) {
	ev.PubKey, err = hex.Dec(s)
	return
}

func (ev *T) SigFromString(s string) (err error) {
	ev.Sig, err = hex.Dec(s)
	return
}

func (ev *T) TagsFromStrings(s ...[]string) {
	ev.Tags = tags.NewWithCap(len(s))
	var tgs []*tag.T
	for _, t := range s {
		tg := tag.New(t...)
		tgs = append(tgs, tg)
	}
	ev.Tags.AppendTags(tgs...)
	return
}

func (ev *T) ContentFromString(s string) {
	ev.Content = []byte(s)
	return
}

// ToEvent converts this above format to the realy native form
func (e J) ToEvent() (ev *T, err error) {
	ev = &T{}
	if err = ev.IDFromString(e.Id); chk.E(err) {
		return
	}
	ev.CreatedAtFromInt64(e.CreatedAt)
	ev.KindFromInt32(e.Kind)
	if err = ev.PubKeyFromString(e.Pubkey); chk.E(err) {
		return
	}
	ev.TagsFromStrings(e.Tags...)
	ev.ContentFromString(e.Content)
	if err = ev.SigFromString(e.Sig); chk.E(err) {
		return
	}
	return
}

func Hash(in []byte) (out []byte) {
	h := sha256.Sum256(in)
	return h[:]
}

func GenerateRandomTextNoteEvent(sign signer.I, maxSize int) (ev *T,
	err error) {

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
