package bunker

import (
	"encoding/json"

	"realy.lol/chk"
	"realy.lol/encryption"
	"realy.lol/event"
	"realy.lol/kind"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

type Session struct {
	Pubkey, SharedKey, ConversationKey []byte
}

func (s *Session) ParseRequest(ev *event.T) (req *Request, err error) {
	var b []byte
	if b, err = encryption.Decrypt(ev.Content, s.ConversationKey); chk.E(err) {
		if b, err = encryption.DecryptNip4(ev.Content, s.SharedKey); chk.E(err) {
			return
		}
	}
	if err = json.Unmarshal(b, &req); chk.E(err) {
		return
	}
	return
}

func (s *Session) MakeResponse(id, requester, result string,
	rErr error) (resp *Response, ev *event.T, err error) {
	if rErr != nil {
		resp = &Response{
			ID:     id,
			Result: rErr.Error(),
		}
	} else if len(result) > 0 {
		resp = &Response{
			ID:     id,
			Result: result,
		}
	}
	// todo: what if the response is empty? this shouldn't happen i think?
	var j []byte
	if j, err = json.Marshal(resp); chk.E(err) {
		return
	}
	var ciphertext []byte
	if ciphertext, err = encryption.Encrypt(j, s.ConversationKey); chk.E(err) {
		return
	}
	ev = &event.T{
		Content:   ciphertext,
		CreatedAt: timestamp.Now(),
		Kind:      kind.NostrConnect,
		Tags:      tags.New(tag.New("p", requester)),
	}
	return
}
