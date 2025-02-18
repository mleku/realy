package httpauth

import (
	"bytes"
	"net/http"
	"testing"

	"realy.lol/lol"
	"realy.lol/p256k"
)

func TestMakeRequest_ValidateRequest(t *testing.T) {
	lol.SetLogLevel("trace")
	sign := new(p256k.Signer)
	err := sign.Generate()
	if chk.E(err) {
		t.Fatal(err)
	}
	var r *http.Request
	if r, err = MakePostRequest("https://example.com/getnpubs?a=b&c=d", "get", sign, ""); chk.E(err) {
		t.Fatal(err)
	}
	var pk []byte
	var valid bool
	if valid, pk, err = ValidateRequest(r); chk.E(err) {
		t.Fatal(err)
	}
	if !valid {
		t.Fatal("request event signature not valid")
	}
	if !bytes.Equal(pk, sign.Pub()) {
		t.Fatalf("unexpected pubkey in nip-98 http auth event: %0x expected %0x",
			pk, sign.Pub())
	}
}