package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"realy.lol/bech32encoding"
	"realy.lol/chk"
	"realy.lol/errorf"
	"realy.lol/httpauth"
	"realy.lol/log"
	"realy.lol/p256k"
	"realy.lol/signer"
)

const secEnv = "NOSTR_SECRET_KEY"

func fail(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func main() {
	// lol.SetLogLevel("trace")
	if len(os.Args) > 1 && os.Args[1] == "help" {
		fmt.Printf(`nauth help:

for generating extended expiration NIP-98 tokens:

    nauth <url prefix> <duration in 0h0m0s format>

	* NIP-98 secret will be expected in the environment variable "%s" - if absent, will not be added to the header. Endpoint is assumed to not require it if absent. An error will be returned if it was needed.

	output will be rendered to stdout

`, secEnv)
		os.Exit(0)
	}
	if len(os.Args) < 3 {
		fail(`error: nauth requires minimum 2 args: <url> <duration in 0h0m0s format>

    signing nsec (in bech32 format) is expected to be found in %s environment variable.

    use "help" to get usage information
`, secEnv)
	}
	ex, err := time.ParseDuration(os.Args[2])
	if err != nil {
		fail(err.Error())
	}
	var sign signer.I
	if sign, err = GetNIP98Signer(); err != nil {
		fail(err.Error())
	}
	exp := time.Now().Add(ex).Unix()
	ev := httpauth.MakeNIP98Event(os.Args[1], "", "", exp)
	if err = ev.Sign(sign); err != nil {
		fail(err.Error())
	}
	log.T.F("nip-98 http auth event:\n%s\n", ev.SerializeIndented())
	b64 := base64.URLEncoding.EncodeToString(ev.Serialize())
	fmt.Println("Nostr " + b64)
}

func GetNIP98Signer() (sign signer.I, err error) {
	nsex := os.Getenv(secEnv)
	var sk []byte
	if len(nsex) == 0 {
		err = errorf.E("no bech32 secret key found in environment variable %s", secEnv)
		return
	} else if sk, err = bech32encoding.NsecToBytes([]byte(nsex)); chk.E(err) {
		err = errorf.E("failed to decode nsec: '%s'", err.Error())
		return
	}
	sign = &p256k.Signer{}
	if err = sign.InitSec(sk); chk.E(err) {
		err = errorf.E("failed to init signer: '%s'", err.Error())
		return
	}
	return
}
