package bunker

import (
	"encoding/json"
	"fmt"
	"sync"

	"relay.mleku.dev/chk"
	"relay.mleku.dev/context"
	"relay.mleku.dev/ec/schnorr"
	"relay.mleku.dev/encryption"
	"relay.mleku.dev/errorf"
	"relay.mleku.dev/event"
	"relay.mleku.dev/hex"
	"relay.mleku.dev/keys"
	"relay.mleku.dev/kind"
	"relay.mleku.dev/signer"
)

var _ Signer = (*StaticKeySigner)(nil)

type StaticKeySigner struct {
	sync.Mutex
	secretKey         signer.I
	sessions          map[string]*Session
	RelaysToAdvertise map[string]RelayReadWrite
	AuthorizeRequest  func(harmless bool, from, secret []byte) bool
}

func NewStaticKeySigner(secretKey signer.I) *StaticKeySigner {
	return &StaticKeySigner{secretKey: secretKey,
		RelaysToAdvertise: make(map[string]RelayReadWrite)}
}

func (p *StaticKeySigner) GetSession(clientPubkey string) (s *Session, exists bool) {
	p.Lock()
	defer p.Unlock()
	s, exists = p.sessions[clientPubkey]
	return
}

func (p *StaticKeySigner) getOrCreateSession(clientPubkey []byte) (s *Session, err error) {
	p.Lock()
	defer p.Unlock()
	s = new(Session)
	var exists bool
	if s, exists = p.sessions[string(clientPubkey)]; exists {
		return
	}
	if s.SharedKey, err = encryption.ComputeSharedSecret(clientPubkey,
		p.secretKey.Sec()); chk.E(err) {
		return
	}
	if s.ConversationKey, err = encryption.GenerateConversationKey(clientPubkey,
		p.secretKey.Pub()); chk.E(err) {
		return
	}
	s.Pubkey = p.secretKey.Pub()
	// add to pool
	p.sessions[string(clientPubkey)] = s
	return
}

func (p *StaticKeySigner) HandleRequest(_ context.T, ev *event.T) (req *Request, res *Response,
	eventResponse *event.T, err error) {
	if !ev.Kind.Equal(kind.NostrConnect) {
		err = errorf.E("event kind is %s, but we expected %s",
			ev.Kind.Name(), kind.NostrConnect.Name())
		return
	}
	var session *Session
	if session, err = p.getOrCreateSession(ev.Pubkey); chk.E(err) {
		return
	}
	if req, err = session.ParseRequest(ev); chk.E(err) {
		return
	}
	var secret, result []byte
	var harmless bool
	var rErr error
	switch req.Method {
	case "connect":
		if len(req.Params) >= 2 {
			secret = req.Params[1]
		}
		result = []byte("ack")
		harmless = true
	case "get_public_key":
		result = session.Pubkey
		harmless = true
	case "sign_event":
		if len(req.Params) != 1 {
			rErr = errorf.E("wrong number of arguments to 'sign_event'")
			break
		}
		evt := &event.T{}
		if rErr = json.Unmarshal(req.Params[0], evt); chk.E(rErr) {
			break
		}
		if rErr = evt.Sign(p.secretKey); chk.E(rErr) {
			break
		}
		result = evt.Serialize()
	case "get_relays":
		if result, rErr = json.Marshal(p.RelaysToAdvertise); chk.E(rErr) {
			break
		}
		harmless = true
	case "nip44_encrypt":
		var pk, sharedSecret []byte
		if pk, rErr = CheckParamsAndKey(req); chk.E(err) {
			break
		}
		if sharedSecret, rErr = p.GetConversationKey(pk); chk.E(err) {
			break
		}
		if result, rErr = encryption.Encrypt(req.Params[1], sharedSecret); chk.E(err) {
			break
		}
	case "nip44_decrypt":
		var pk, sharedSecret []byte
		if pk, rErr = CheckParamsAndKey(req); chk.E(err) {
			break
		}
		if sharedSecret, rErr = p.GetConversationKey(pk); chk.E(err) {
			break
		}
		if result, err = encryption.Decrypt(req.Params[1], sharedSecret); chk.E(err) {
			break
		}
	case "nip04_encrypt":
		var pk, sharedSecret []byte
		if pk, rErr = CheckParamsAndKey(req); chk.E(err) {
			break
		}
		if sharedSecret, rErr = p.ComputeSharedSecret(pk); chk.E(err) {
			break
		}
		if result, rErr = encryption.EncryptNip4(req.Params[1],
			sharedSecret); chk.E(err) {
			break
		}
	case "nip04_decrypt":
		var pk, sharedSecret []byte
		if pk, rErr = CheckParamsAndKey(req); chk.E(err) {
			break
		}
		if sharedSecret, rErr = p.ComputeSharedSecret(pk); chk.E(err) {
			break
		}
		if result, rErr = encryption.DecryptNip4(req.Params[1],
			sharedSecret); chk.E(err) {
			break
		}
	case "ping":
		result = []byte("pong")
		harmless = true
	default:
		rErr = errorf.E("unknown method '%s'", req.Method)
	}
	if rErr == nil && p.AuthorizeRequest != nil {
		if !p.AuthorizeRequest(harmless, ev.Pubkey, secret) {
			rErr = fmt.Errorf("unauthorized")
		}
	}
	if res, eventResponse, err = session.MakeResponse(req.ID, hex.Enc(ev.Pubkey),
		string(result), rErr); chk.E(err) {
		return
	}
	if err = eventResponse.Sign(p.secretKey); chk.E(err) {
		return
	}
	return
}

func (p *StaticKeySigner) GetConversationKey(pk []byte) (sharedSecret []byte, rErr error) {
	if sharedSecret, rErr = encryption.GenerateConversationKey(pk,
		p.secretKey.Sec()); chk.E(rErr) {
		return
	}
	return
}

func (p *StaticKeySigner) ComputeSharedSecret(pk []byte) (sharedSecret []byte, rErr error) {
	if sharedSecret, rErr = encryption.ComputeSharedSecret(pk,
		p.secretKey.Sec()); chk.E(rErr) {
		return
	}
	return
}

func CheckParamsAndKey(req *Request) (pk []byte, rErr error) {
	if len(req.Params) != 2 {
		rErr = errorf.E("wrong number of arguments to 'nip04_decrypt'")
		return
	}
	if !keys.IsValidPublicKey(req.Params[0]) {
		rErr = errorf.E("first argument to 'nip04_decrypt' is not a pubkey string")
		return
	}
	pk = make([]byte, schnorr.PubKeyBytesLen)
	if _, rErr = hex.DecBytes(pk, req.Params[0]); chk.E(rErr) {
		return
	}
	return
}
