package event

import (
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
	jId        = by("id")
	jPubkey    = by("pubkey")
	jCreatedAt = by("created_at")
	jKind      = by("kind")
	jTags      = by("tags")
	jContent   = by("content")
	jSig       = by("sig")
)

func (ev *T) Marshal(dst by) (b by) {
	// open parentheses
	dst = append(dst, '{')
	// ID
	dst = text.JSONKey(dst, jId)
	dst = text.AppendQuote(dst, ev.ID, hex.EncAppend)
	dst = append(dst, ',')
	// PubKey
	dst = text.JSONKey(dst, jPubkey)
	dst = text.AppendQuote(dst, ev.PubKey, hex.EncAppend)
	dst = append(dst, ',')
	// CreatedAt
	dst = text.JSONKey(dst, jCreatedAt)
	dst = ev.CreatedAt.Marshal(dst)
	dst = append(dst, ',')
	// Kind
	dst = text.JSONKey(dst, jKind)
	dst = ev.Kind.Marshal(dst)
	dst = append(dst, ',')
	// Tags
	dst = text.JSONKey(dst, jTags)
	dst = ev.Tags.Marshal(dst)
	dst = append(dst, ',')
	// Content
	dst = text.JSONKey(dst, jContent)
	dst = text.AppendQuote(dst, ev.Content, text.NostrEscape)
	dst = append(dst, ',')
	// jSig
	dst = text.JSONKey(dst, jSig)
	dst = text.AppendQuote(dst, ev.Sig, hex.EncAppend)
	// close parentheses
	dst = append(dst, '}')
	b = dst
	return
}

func Marshal(ev *T, dst by) (b by) { return ev.Marshal(dst) }

func (ev *T) Unmarshal(b by) (r by, err er) {
	key := make(by, 0, 9)
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
		if !equals(jId, key) {
			goto invalid
		}
		var id by
		if id, r, err = text.UnmarshalHex(r); chk.E(err) {
			return
		}
		if len(id) != sha256.Size {
			err = errorf.E("invalid ID, require %d got %d", sha256.Size,
				len(id))
			return
		}
		ev.ID = id
		goto BetweenKV
	case jPubkey[0]:
		if !equals(jPubkey, key) {
			goto invalid
		}
		var pk by
		if pk, r, err = text.UnmarshalHex(r); chk.E(err) {
			return
		}
		if len(pk) != schnorr.PubKeyBytesLen {
			err = errorf.E("invalid pubkey, require %d got %d",
				schnorr.PubKeyBytesLen, len(pk))
			return
		}
		ev.PubKey = pk
		goto BetweenKV
	case jKind[0]:
		if !equals(jKind, key) {
			goto invalid
		}
		ev.Kind = kind.New(0)
		if r, err = ev.Kind.Unmarshal(r); chk.E(err) {
			return
		}
		goto BetweenKV
	case jTags[0]:
		if !equals(jTags, key) {
			goto invalid
		}
		ev.Tags = tags.New()
		if r, err = ev.Tags.Unmarshal(r); chk.E(err) {
			return
		}
		goto BetweenKV
	case jSig[0]:
		if !equals(jSig, key) {
			goto invalid
		}
		var sig by
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
			if !equals(jContent, key) {
				goto invalid
			}
			if ev.Content, r, err = text.UnmarshalQuoted(r); chk.T(err) {
				return
			}
			goto BetweenKV
		} else if key[1] == jCreatedAt[1] {
			if !equals(jCreatedAt, key) {
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
	err = errorf.E("invalid key,\n'%s'\n'%s'\n'%s'", st(b), st(b[:len(r)]),
		st(r))
	return
eof:
	err = io.EOF
	return
}

func Unmarshal(ev *T, b by) (r by, err er) { return ev.Unmarshal(b) }
