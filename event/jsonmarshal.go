package event

import (
	"realy.lol/hex"
	"realy.lol/text"
)

func (ev *T) MarshalJSON(dst B) (b B, err error) {
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
	if dst, err = ev.CreatedAt.MarshalJSON(dst); chk.E(err) {
		return
	}
	dst = append(dst, ',')
	// Kind
	dst = text.JSONKey(dst, jKind)
	dst, _ = ev.Kind.MarshalJSON(dst)
	dst = append(dst, ',')
	// Tags
	dst = text.JSONKey(dst, jTags)
	dst, _ = ev.Tags.MarshalJSON(dst)
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
