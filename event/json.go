package event

import (
	"bytes"
	"encoding/json"
	"io"

	"realy.lol/ec/schnorr"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/sha256"
	"realy.lol/tags"
	"realy.lol/text"
	"realy.lol/timestamp"
)

var (
	jId        = []byte("id")
	jPubkey    = []byte("pubkey")
	jCreatedAt = []byte("created_at")
	jKind      = []byte("kind")
	jTags      = []byte("tags")
	jContent   = []byte("content")
	jSig       = []byte("sig")
)

// Marshal appends an event.T to a provided destination slice.
func (ev *T) Marshal(dst []byte) (b []byte) {
	b = ev.marshalWithWhitespace(dst, false)
	return
}

// marshalWithWhitespace adds tabs and newlines to make the JSON more readable
// for humans, if the on flag is set to true.
func (ev *T) marshalWithWhitespace(dst []byte, on bool) (b []byte) {
	// open parentheses
	dst = append(dst, '{')
	// Id
	if on {
		dst = append(dst, '\n', '\t')
	}
	dst = text.JSONKey(dst, jId)
	dst = text.AppendQuote(dst, ev.Id, hex.EncAppend)
	dst = append(dst, ',')
	// Pubkey
	if on {
		dst = append(dst, '\n', '\t')
	}
	dst = text.JSONKey(dst, jPubkey)
	dst = text.AppendQuote(dst, ev.Pubkey, hex.EncAppend)
	dst = append(dst, ',')
	if on {
		dst = append(dst, '\n', '\t')
	}
	// CreatedAt
	dst = text.JSONKey(dst, jCreatedAt)
	dst = ev.CreatedAt.Marshal(dst)
	dst = append(dst, ',')
	if on {
		dst = append(dst, '\n', '\t')
	}
	// Kind
	dst = text.JSONKey(dst, jKind)
	dst = ev.Kind.Marshal(dst)
	dst = append(dst, ',')
	if on {
		dst = append(dst, '\n', '\t')
	}
	// Tags
	dst = text.JSONKey(dst, jTags)
	dst = ev.Tags.Marshal(dst)
	dst = append(dst, ',')
	if on {
		dst = append(dst, '\n', '\t')
	}
	// Content
	dst = text.JSONKey(dst, jContent)
	dst = text.AppendQuote(dst, ev.Content, text.NostrEscape)
	dst = append(dst, ',')
	if on {
		dst = append(dst, '\n', '\t')
	}
	// jSig
	dst = text.JSONKey(dst, jSig)
	dst = text.AppendQuote(dst, ev.Sig, hex.EncAppend)
	if on {
		dst = append(dst, '\n')
	}
	// close parentheses
	dst = append(dst, '}')
	b = dst
	return
}

// Marshal is a normal function that is the same as event.T Marshal method
// except you explicitly specify the receiver.
func Marshal(ev *T, dst []byte) (b []byte) { return ev.Marshal(dst) }

// Unmarshal an event from minified JSON into an event.T.
func (ev *T) Unmarshal(b []byte) (r []byte, err error) {
	// this parser does not cope with whitespaces in valid places in json, so we
	// scan first for linebreaks, as these indicate that it is probably not gona work and fall back to json.Unmarshal
	for _, v := range b {
		if v == '\n' {
			// revert to json.Unmarshal
			var j J
			if err = json.Unmarshal(b, &j); chk.E(err) {
				return
			}
			var e *T
			if e, err = j.ToEvent(); chk.E(err) {
				return
			}
			*ev = *e
			return
		}
	}

	key := make([]byte, 0, 9)
	r = b
	for ; len(r) > 0; r = r[1:] {
		if r[0] == '{' {
			r = r[1:]
			goto BetweenKeys
		}
	}
	goto eof
BetweenKeys:
	for ; len(r) > 0; r = r[1:] {
		if r[0] == '"' {
			r = r[1:]
			goto InKey
		}
	}
	goto eof
InKey:
	for ; len(r) > 0; r = r[1:] {
		if r[0] == '"' {
			r = r[1:]
			goto InKV
		}
		key = append(key, r[0])
	}
	goto eof
InKV:
	for ; len(r) > 0; r = r[1:] {
		if r[0] == ':' {
			r = r[1:]
			goto InVal
		}
	}
	goto eof
InVal:
	switch key[0] {
	case jId[0]:
		if !bytes.Equal(jId, key) {
			goto invalid
		}
		var id []byte
		if id, r, err = text.UnmarshalHex(r); chk.E(err) {
			return
		}
		if len(id) != sha256.Size {
			err = errorf.E("invalid Id, require %d got %d", sha256.Size,
				len(id))
			return
		}
		ev.Id = id
		goto BetweenKV
	case jPubkey[0]:
		if !bytes.Equal(jPubkey, key) {
			goto invalid
		}
		var pk []byte
		if pk, r, err = text.UnmarshalHex(r); chk.E(err) {
			return
		}
		if len(pk) != schnorr.PubKeyBytesLen {
			err = errorf.E("invalid pubkey, require %d got %d",
				schnorr.PubKeyBytesLen, len(pk))
			return
		}
		ev.Pubkey = pk
		goto BetweenKV
	case jKind[0]:
		if !bytes.Equal(jKind, key) {
			goto invalid
		}
		ev.Kind = kind.New(0)
		if r, err = ev.Kind.Unmarshal(r); chk.E(err) {
			return
		}
		goto BetweenKV
	case jTags[0]:
		if !bytes.Equal(jTags, key) {
			goto invalid
		}
		ev.Tags = tags.New()
		if r, err = ev.Tags.Unmarshal(r); chk.E(err) {
			return
		}
		goto BetweenKV
	case jSig[0]:
		if !bytes.Equal(jSig, key) {
			goto invalid
		}
		var sig []byte
		if sig, r, err = text.UnmarshalHex(r); chk.E(err) {
			return
		}
		if len(sig) != schnorr.SignatureSize {
			err = errorf.E("invalid sig length, require %d got %d '%s'\n%s",
				schnorr.SignatureSize, len(sig), r, b)
			return
		}
		ev.Sig = sig
		goto BetweenKV
	case jContent[0]:
		if key[1] == jContent[1] {
			if !bytes.Equal(jContent, key) {
				goto invalid
			}
			if ev.Content, r, err = text.UnmarshalQuoted(r); chk.T(err) {
				return
			}
			goto BetweenKV
		} else if key[1] == jCreatedAt[1] {
			if !bytes.Equal(jCreatedAt, key) {
				goto invalid
			}
			ev.CreatedAt = timestamp.New()
			if r, err = ev.CreatedAt.Unmarshal(r); chk.T(err) {
				return
			}
			goto BetweenKV
		} else {
			goto invalid
		}
	default:
		goto invalid
	}
BetweenKV:
	key = key[:0]
	for ; len(r) > 0; r = r[1:] {
		switch {
		case len(r) == 0:
			return
		case r[0] == '}':
			r = r[1:]
			goto AfterClose
		case r[0] == ',':
			r = r[1:]
			goto BetweenKeys
		case r[0] == '"':
			r = r[1:]
			goto InKey
		}
	}
	goto eof
AfterClose:
	return
invalid:
	err = errorf.E("invalid key,\n'%s'\n'%s'\n'%s'", string(b), string(b[:len(r)]),
		string(r))
	return
eof:
	err = io.EOF
	return
}

// Unmarshal is the same as the event.T Unmarshal method except you give it the
// event to marshal into instead of call it as a method of the type.
func Unmarshal(ev *T, b []byte) (r []byte, err error) { return ev.Unmarshal(b) }
