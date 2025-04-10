// Package event provides a codec for nostr events, for the wire format (with Id
// and signature), for the canonical form, that is hashed to generate the Id,
// and a custom format called "wirecompact" which wraps a canonical form with an
// array and encodes the signature with base64 for a more compact size.
package event

import (
	"lukechampine.com/frand"

	"realy.mleku.dev/ec/schnorr"
	"realy.mleku.dev/eventid"
	"realy.mleku.dev/hex"
	"realy.mleku.dev/kind"
	"realy.mleku.dev/sha256"
	"realy.mleku.dev/signer"
	"realy.mleku.dev/tag"
	"realy.mleku.dev/tags"
	"realy.mleku.dev/text"
	"realy.mleku.dev/timestamp"
	"realy.mleku.dev/unix"
)

// T is the primary datatype of nostr. This is the form of the structure that
// defines its JSON string based format.
type T struct {

	// Id is the SHA256 hash of the canonical encoding of the event in binary format
	Id []byte

	// Pubkey is the public key of the event creator in binary format
	Pubkey []byte

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

	// Sig is the signature on the Id hash that validates as coming from the
	// Pubkey in binary format.
	Sig []byte
}

// Ts is an array of event.T that sorts in reverse chronological order.
type Ts []*T

// Len returns the length of the event.Ts.
func (ev Ts) Len() int { return len(ev) }

// Less returns whether the first is newer than the second (larger unix
// timestamp).
func (ev Ts) Less(i, j int) bool { return ev[i].CreatedAt.I64() > ev[j].CreatedAt.I64() }

// Swap two indexes of the event.Ts with each other.
func (ev Ts) Swap(i, j int) { ev[i], ev[j] = ev[j], ev[i] }

// C is a channel that carries event.T.
type C chan *T

// New makes a new event.T.
func New() (ev *T) { return &T{} }

// Serialize renders an event.T into minified JSON.
func (ev *T) Serialize() (b []byte) { return ev.Marshal(nil) }

// SerializeIndented renders an event.T into nicely readable whitespaced JSON.
func (ev *T) SerializeIndented() (b []byte) {
	return ev.marshalWithWhitespace(nil, true)
}

// EventId returns the event.T Id as an eventid.T.
func (ev *T) EventId() (eid *eventid.T) {
	return eventid.NewWith(ev.Id)
}

// stringy/numbery functions for retarded other libraries

// IdString returns the event Id as a hex encoded string.
func (ev *T) IdString() (s string) { return hex.Enc(ev.Id) }

// CreatedAtInt64 returns the created_at timestamp as a standard int64.
func (ev *T) CreatedAtInt64() (i int64) { return ev.CreatedAt.I64() }

// KindInt32 returns the kind as an int32, as is often needed for JSON.
func (ev *T) KindInt32() (i int32) { return int32(ev.Kind.K) }

// PubKeyString returns the pubkey as a hex encoded string.
func (ev *T) PubKeyString() (s string) { return hex.Enc(ev.Pubkey) }

// SigString returns the signature as a hex encoded string.
func (ev *T) SigString() (s string) { return hex.Enc(ev.Sig) }

// TagStrings returns the tags as a slice of slice of strings.
func (ev *T) TagStrings() (s [][]string) { return ev.Tags.ToStringsSlice() }

// ContentString returns the content field as a string.
func (ev *T) ContentString() (s string) { return string(ev.Content) }

// J is an event.T encoded in more basic types than used in this library.
type J struct {
	Id        string     `json:"id"`
	Pubkey    string     `json:"pubkey"`
	CreatedAt unix.Time  `json:"created_at"`
	Kind      int32      `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

// ToEventJ converts an event.T into an event.J.
func (ev *T) ToEventJ() (j *J) {
	j = &J{}
	j.Id = ev.IdString()
	j.Pubkey = ev.PubKeyString()
	j.CreatedAt = unix.Time{ev.CreatedAt.Time()}
	j.Kind = ev.KindInt32()
	j.Content = ev.ContentString()
	j.Tags = ev.Tags.ToStringsSlice()
	j.Sig = ev.SigString()
	return
}

// IdFromString decodes an event ID and loads it into an event.T Id.
func (ev *T) IdFromString(s string) (err error) {
	ev.Id, err = hex.Dec(s)
	return
}

// CreatedAtFromInt64 encodes a unix timestamp into the CreatedAt field of an
// event.T.
func (ev *T) CreatedAtFromInt64(i int64) {
	ev.CreatedAt = timestamp.FromUnix(i)
	return
}

// KindFromInt32 encodes an int32 representation of a kind.T into an event.T.
func (ev *T) KindFromInt32(i int32) {
	ev.Kind = &kind.T{}
	ev.Kind.K = uint16(i)
	return
}

// PubKeyFromString decodes a hex-encoded string into the event.T Pubkey field.
func (ev *T) PubKeyFromString(s string) (err error) {
	if len(s) != 2*schnorr.PubKeyBytesLen {
		err = errorf.E("invalid length public key hex, got %d require %d",
			len(s), 2*schnorr.PubKeyBytesLen)
	}
	ev.Pubkey, err = hex.Dec(s)
	return
}

// SigFromString decodes a hex-encoded string into the event.T Sig field.
func (ev *T) SigFromString(s string) (err error) {
	if len(s) != 2*schnorr.SignatureSize {
		err = errorf.E("invalid length signature hex, got %d require %d",
			len(s), 2*schnorr.SignatureSize)
	}
	ev.Sig, err = hex.Dec(s)
	return
}

// TagsFromStrings converts a slice of slice of strings into a tags.T for the
// event.T.
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

// ContentFromString imports a content string into the event.T Content field.
func (ev *T) ContentFromString(s string) {
	ev.Content = []byte(s)
	return
}

// ToEvent converts event.J format to the realy native form.
func (e J) ToEvent() (ev *T, err error) {
	ev = &T{}
	if err = ev.IdFromString(e.Id); chk.E(err) {
		return
	}
	ev.CreatedAtFromInt64(e.CreatedAt.Unix())
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

// Hash is a little helper generate a hash and return a slice instead of an
// array.
func Hash(in []byte) (out []byte) {
	h := sha256.Sum256(in)
	return h[:]
}

// GenerateRandomTextNoteEvent creates a generic event.T with random text
// content.
func GenerateRandomTextNoteEvent(sign signer.I, maxSize int) (ev *T,
	err error) {

	l := frand.Intn(maxSize * 6 / 8) // account for base64 expansion
	ev = &T{
		Pubkey:    sign.Pub(),
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
