package httpauth

import (
	"testing"
)

func TestMakeNIP98Request_ValidateNIP98Request(t *testing.T) {
	// lol.SetLogLevel("trace")
	// sign := new(p256k.Signer)
	// err := sign.Generate()
	// if chk.E(err) {
	// 	t.Fatal(err)
	// }
	// // var ur *url.URL
	// // if ur, err = url.Parse("https://example.com/getnpubs?a=b&c=d"); chk.E(err) {
	// // 	t.Fatal(err)
	// // }
	// var r *http.Request
	// // if r, err = MakeNIP98GetRequest(ur, "test/0.0.0", sign); chk.E(err) {
	// // 	t.Fatal(err)
	// // }
	// var pk []byte
	// var valid bool
	// if valid, pk, err = CheckAuth(r, nil); chk.E(err) {
	// 	t.Fatal(err)
	// }
	// if !valid {
	// 	t.Fatal("request event signature not valid")
	// }
	// if !bytes.Equal(pk, sign.Pub()) {
	// 	t.Fatalf("unexpected pubkey in nip-98 http auth event: %0x expected %0x",
	// 		pk, sign.Pub())
	// }
}
