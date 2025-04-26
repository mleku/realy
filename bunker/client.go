package bunker

import (
	"encoding/json"
	"math/rand"
	"net/url"
	"strconv"

	"github.com/puzpuzpuz/xsync/v3"

	"relay.mleku.dev/atomic"
	"relay.mleku.dev/chk"
	"relay.mleku.dev/context"
	"relay.mleku.dev/encryption"
	"relay.mleku.dev/errorf"
	"relay.mleku.dev/event"
	"relay.mleku.dev/filter"
	"relay.mleku.dev/filters"
	"relay.mleku.dev/hex"
	"relay.mleku.dev/keys"
	"relay.mleku.dev/kind"
	"relay.mleku.dev/kinds"
	"relay.mleku.dev/p256k"
	"relay.mleku.dev/signer"
	"relay.mleku.dev/tag"
	"relay.mleku.dev/tags"
	"relay.mleku.dev/timestamp"
	"relay.mleku.dev/ws"
)

type BunkerClient struct {
	serial          atomic.Uint64
	clientSecretKey signer.I
	target          []byte
	pool            *ws.Pool
	relays          []string
	conversationKey []byte // nip44
	listeners       *xsync.MapOf[string, chan Response]
	expectingAuth   *xsync.MapOf[string, struct{}]
	idPrefix        string
	onAuth          func(string)
	// memoized
	getPublicKeyResponse string
	// SkipSignatureCheck can be set if you don't want to double-check incoming signatures
	SkipSignatureCheck bool
}

// ConnectBunker establishes an RPC connection to a NIP-46 signer using the relays and secret provided in the bunkerURL.
// pool can be passed to reuse an existing pool, otherwise a new pool will be created.
func ConnectBunker(
	ctx context.T,
	clientSecretKey signer.I,
	bunkerURLOrNIP05 string,
	pool *ws.Pool,
	onAuth func(string),
) (client *BunkerClient, err error) {
	var parsed *url.URL
	if parsed, err = url.Parse(bunkerURLOrNIP05); chk.E(err) {
		return
	}
	// assume it's a bunker url (will fail later if not)
	secret := parsed.Query().Get("secret")
	relays := parsed.Query()["relay"]
	targetPublicKey := parsed.Host
	if parsed.Scheme == "" {
		// could be a NIP-05
		var pubkey string
		var relays_ []string
		if pubkey, relays_, err = queryWellKnownNostrJson(ctx, bunkerURLOrNIP05); chk.E(err) {
			return
		}
		targetPublicKey = pubkey
		relays = relays_
	} else if parsed.Scheme == "bunker" {
		// this is what we were expecting, so just move on
	} else {
		// otherwise fail here
		err = errorf.E("wrong scheme '%s', must be bunker://", parsed.Scheme)
		return
	}
	if !keys.IsValidPublicKey(targetPublicKey) {
		err = errorf.E("'%s' is not a valid public key hex", targetPublicKey)
		return
	}
	var targetPubkey, sec []byte
	if targetPubkey, err = keys.HexPubkeyToBytes(targetPublicKey); chk.E(err) {
		return
	}
	if sec, err = hex.Dec(secret); chk.E(err) {
		return
	}
	if client, err = NewBunker(
		ctx,
		clientSecretKey,
		targetPubkey,
		relays,
		pool,
		onAuth,
	); chk.E(err) {
		return
	}
	_, err = client.RPC(ctx, "connect", [][]byte{targetPubkey, sec})
	return
}

func NewBunker(
	ctx context.T,
	clientSecretKey signer.I,
	targetPublicKey []byte,
	relays []string,
	pool *ws.Pool,
	onAuth func(string),
) (client *BunkerClient, err error) {
	if pool == nil {
		pool = ws.NewPool(ctx)
	}
	clientSecret := new(p256k.Signer)
	if err = clientSecret.InitSec(clientSecretKey.Sec()); chk.E(err) {
		return
	}
	clientPubkey := clientSecret.Pub()
	var conversationKey, sharedSecret []byte
	if sharedSecret, err = encryption.ComputeSharedSecret(targetPublicKey,
		clientSecretKey.Sec()); chk.E(err) {
		return
	}
	if conversationKey, err = encryption.GenerateConversationKey(targetPublicKey,
		clientSecret.Sec()); chk.E(err) {
		return
	}
	client = &BunkerClient{
		pool:            pool,
		clientSecretKey: clientSecretKey,
		target:          targetPublicKey,
		relays:          relays,
		conversationKey: conversationKey,
		listeners:       xsync.NewMapOf[string, chan Response](),
		expectingAuth:   xsync.NewMapOf[string, struct{}](),
		onAuth:          onAuth,
		idPrefix:        "gn-" + strconv.Itoa(rand.Intn(65536)),
	}
	go func() {
		now := timestamp.Now()
		events := pool.SubMany(ctx, relays, filters.New(&filter.T{
			Tags:  tags.New(tag.New([]byte("p"), clientPubkey)),
			Kinds: kinds.New(kind.NostrConnect),
			Since: now,
		}), ws.WithLabel("bunker46client"))
		for ev := range events {
			if !ev.Event.Kind.Equal(kind.NostrConnect) {
				err = errorf.E("event kind is %s, but we expected %s",
					ev.Event.Kind.Name(), kind.NostrConnect.Name())
				continue
			}
			var plain []byte
			if plain, err = encryption.Decrypt(ev.Event.Content, conversationKey); chk.E(err) {
				if plain, err = encryption.DecryptNip4(ev.Event.Content,
					sharedSecret); chk.E(err) {
					continue
				}
			}
			var resp Response
			if err = json.Unmarshal(plain, &resp); chk.E(err) {
				continue
			}
			if resp.Result == "auth_url" {
				// special case
				authURL := resp.Error
				if _, ok := client.expectingAuth.Load(resp.ID); ok {
					client.onAuth(authURL)
				}
				continue
			}
			if dispatcher, ok := client.listeners.Load(resp.ID); ok {
				dispatcher <- resp
				continue
			}
		}
	}()
	return
}

func (client *BunkerClient) RPC(ctx context.T, method string,
	params [][]byte) (result string, err error) {
	id := client.idPrefix + "-" + strconv.FormatUint(client.serial.Add(1), 10)
	var req []byte
	if req, err = json.Marshal(Request{
		ID:     id,
		Method: method,
		Params: params,
	}); chk.E(err) {
		return
	}
	var content []byte
	if content, err = encryption.Encrypt(req, client.conversationKey); chk.E(err) {
		return
	}
	ev := &event.T{
		Content:   content,
		CreatedAt: timestamp.Now(),
		Kind:      kind.NostrConnect,
		Tags:      tags.New(tag.New([]byte("p"), client.target)),
	}
	if err = ev.Sign(client.clientSecretKey); chk.E(err) {
		return
	}
	respWaiter := make(chan Response)
	client.listeners.Store(id, respWaiter)
	defer func() {
		client.listeners.Delete(id)
		close(respWaiter)
	}()
	hasWorked := make(chan struct{})
	for _, url := range client.relays {
		go func(url string) {
			var relay *ws.Client
			relay, err = client.pool.EnsureRelay(url)
			if err == nil {
				if err = relay.Publish(ctx, ev); chk.E(err) {
					return
				}
				select {
				case hasWorked <- struct{}{}: // todo: shouldn't this be after success?
				default:
				}
			}
		}(url)
	}
	select {
	case <-hasWorked:
		// continue
	case <-ctx.Done():
		err = errorf.E("couldn't connect to any relay")
		return
	}

	select {
	case <-ctx.Done():
		err = errorf.E("context canceled")
		return
	case resp := <-respWaiter:
		if resp.Error != "" {
			err = errorf.E("response error: %s", resp.Error)
			return
		}
		result = resp.Result
		return
	}
}

func (client *BunkerClient) Ping(ctx context.T) (err error) {
	if _, err = client.RPC(ctx, "ping", [][]byte{}); chk.E(err) {
		return
	}
	return
}

func (client *BunkerClient) GetPublicKey(ctx context.T) (resp string, err error) {
	if client.getPublicKeyResponse != "" {
		resp = client.getPublicKeyResponse
		return
	}
	resp, err = client.RPC(ctx, "get_public_key", [][]byte{})
	client.getPublicKeyResponse = resp
	return
}

func (client *BunkerClient) SignEvent(ctx context.T, evt *event.T) (err error) {
	var resp string
	if resp, err = client.RPC(ctx, "sign_event", [][]byte{evt.Serialize()}); chk.E(err) {
		return
	}
	if err = json.Unmarshal([]byte(resp), evt); chk.E(err) {
		return
	}
	if !client.SkipSignatureCheck {
		var valid bool
		if valid, err = evt.Verify(); chk.E(err) {
			return
		}
		if !valid {
			err = errorf.E("sign_event response from bunker has invalid signature")
			return
		}
	}
	return
}

func (client *BunkerClient) NIP44Encrypt(ctx context.T,
	targetPublicKey, plaintext []byte) (string, error) {
	return client.RPC(ctx, "nip44_encrypt", [][]byte{targetPublicKey, plaintext})
}

func (client *BunkerClient) NIP44Decrypt(ctx context.T,
	targetPublicKey, ciphertext []byte) (string, error) {
	return client.RPC(ctx, "nip44_decrypt", [][]byte{targetPublicKey, ciphertext})
}

func (client *BunkerClient) NIP04Encrypt(ctx context.T,
	targetPublicKey, plaintext []byte) (string, error) {
	return client.RPC(ctx, "nip04_encrypt", [][]byte{targetPublicKey, plaintext})
}

func (client *BunkerClient) NIP04Decrypt(ctx context.T,
	targetPublicKey, ciphertext []byte) (string, error) {
	return client.RPC(ctx, "nip04_decrypt", [][]byte{targetPublicKey, ciphertext})
}
