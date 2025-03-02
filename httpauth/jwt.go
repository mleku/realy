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
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

const MaxSkew = 15

type JWT struct {
	Issuer         string `json:"iss"`
	Subject        string `json:"sub"`
	Algorithm      string `json:"alg"`
	IssuedAt       int64  `json:"iat"`
	ExpirationTime int64  `json:"exp,omitempty"`
	NotBefore      int64  `json:"nbf,omitempty"`
	Audience       string `json:"aud,omitempty"`
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
		claim.ExpirationTime = claim.IssuedAt + int64(dur/time.Second)
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
	if headerEntry, err = token.SignedString(sec); chk.E(err) {
		return
	}
	return
}

func VerifyJWTtoken(entry, URL, npub, jwtPub string) (valid bool, err error) {

	var token *jwt.Token
	if token, err = jwt.Parse(entry, func(token *jwt.Token) (ifc interface{}, err error) {
		var pkb []byte
		if pkb, err = base64.URLEncoding.DecodeString(jwtPub); chk.E(err) {
			return
		}
		var jpk any
		if jpk, err = x509.ParsePKIXPublicKey(pkb); chk.E(err) {
			return
		}
		ifc = jpk
		var sub string
		if sub, err = token.Claims.GetSubject(); sub != URL {
			err = errors.Wrap(jwt.ErrTokenInvalidClaims, "subject doesn't match expected URL")
			return
		}
		now := time.Now().Unix()
		var exp *jwt.NumericDate
		if exp, err = token.Claims.GetExpirationTime(); chk.E(err) {
		}
		if exp != nil {
			cmp := now - exp.Unix()
			if cmp > MaxSkew {
				err = errors.Wrapf(jwt.ErrTokenInvalidClaims,
					"token is expired, %ds since expiry %d, time now %d, max allowed %d", cmp, exp.Unix(), now, MaxSkew)
				return
			}
		} else {
			var iat *jwt.NumericDate
			if iat, err = token.Claims.GetIssuedAt(); chk.E(err) {
				return
			}
			cmp := time.Now().Unix() - iat.Unix()
			if cmp > 15 || cmp < -15 {
				err = errors.Wrapf(jwt.ErrTokenInvalidClaims,
					"issued at is more than %d seconds skewed", cmp)
				return
			}
		}
		var iss string
		if iss, err = token.Claims.GetIssuer(); chk.E(err) {
			return
		}
		if iss != npub {
			err = errors.Wrapf(jwt.ErrTokenInvalidClaims, "expected issuer %s, got %s", npub, iss)
			return
		}
		return
	}, jwt.WithoutClaimsValidation()); chk.E(err) {
		return
	}
	valid = token.Valid
	return
}
