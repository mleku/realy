package httpauth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWT struct {
	Issuer    string `json:"iss"`
	Type      string `json:"typ"`
	Subject   string `json:"sub"`
	Algorithm string `json:"alg"`
	IssuedAt  int64  `json:"iat"`
	Expiry    int64  `json:"exp,omitempty"`
}

func GenerateJWTKeys() (x509sec, x509pub, pemSec, pemPub []byte, err error) {
	var sec *ecdsa.PrivateKey
	if sec, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); chk.E(err) {
		return
	}
	var pkb []byte
	if pkb, err = x509.MarshalPKIXPublicKey(sec.Public()); chk.E(err) {
		return
	}
	x509pub = make([]byte, len(pkb)*8/6+3)
	base64.URLEncoding.Encode(x509pub, pkb)

	var skb []byte
	if skb, err = x509.MarshalECPrivateKey(sec); chk.E(err) {
		return
	}
	x509sec = make([]byte, len(skb)*8/6+3)
	base64.URLEncoding.Encode(x509sec, skb)

	bufS := new(bytes.Buffer)
	if err = pem.Encode(bufS, &pem.Block{"EC PRIVATE KEY",
		nil, skb}); chk.E(err) {
		return
	}
	pemSec = bufS.Bytes()

	bufP := new(bytes.Buffer)
	if err = pem.Encode(bufP, &pem.Block{"EC PUBLIC KEY",
		nil, pkb}); chk.E(err) {
		return
	}
	pemPub = bufP.Bytes()
	return
}

func GenerateJWTtoken(issuer, ur string,
	exp ...string) (b []byte, err error) {
	// generate claim
	claim := &JWT{
		Issuer:    issuer,
		Type:      "message",
		Subject:   ur,
		Algorithm: "ES256",
		IssuedAt:  time.Now().Unix(),
	}
	if len(exp) > 0 {
		// parse duration
		var dur time.Duration
		if dur, err = time.ParseDuration(exp[0]); chk.E(err) {
			return
		}
		claim.Expiry = claim.IssuedAt + int64(dur/time.Second)
	}
	if b, err = json.Marshal(claim); chk.E(err) {
		return
	}
	return
}

func SignJWTtoken(tok []byte, sec *ecdsa.PrivateKey) (headerEntry string, err error) {
	var claims jwt.MapClaims
	if err = json.Unmarshal(tok, &claims); chk.E(err) {
		return
	}
	alg := jwt.GetSigningMethod(claims["alg"].(string))
	token := jwt.NewWithClaims(alg, claims)
	var signed string
	if signed, err = token.SignedString(sec); chk.E(err) {
		return
	}
	headerEntry = fmt.Sprintf("Authorization: Bearer %s\n", signed)
	return
}

func VerifyJWTtoken() {

}
