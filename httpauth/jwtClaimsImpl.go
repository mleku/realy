package httpauth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (j *JWT) GetExpirationTime() (exp *jwt.NumericDate, err error) {
	exp = jwt.NewNumericDate(time.Unix(j.ExpirationTime, 0))
	return
}

func (j *JWT) GetIssuedAt() (iat *jwt.NumericDate, err error) {
	iat = jwt.NewNumericDate(time.Unix(j.IssuedAt, 0))
	return
}

func (j *JWT) GetNotBefore() (nbf *jwt.NumericDate, err error) {
	nbf = jwt.NewNumericDate(time.Unix(j.NotBefore, 0))
	return
}

func (j *JWT) GetIssuer() (iss string, err error) {
	iss = j.Issuer
	return
}

func (j *JWT) GetSubject() (sub string, err error) {
	sub = j.Subject
	return
}

func (j *JWT) GetAudience() (aud jwt.ClaimStrings, err error) {
	aud = jwt.ClaimStrings{j.Audience}
	return
}

func (j *JWT) Validate() (err error) {
	log.I.S("validate")
	return
}
