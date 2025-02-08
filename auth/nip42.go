package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"strings"
	"time"

	"realy.lol/event"
	"realy.lol/kind"
	"realy.lol/tag"
	"realy.lol/tags"
	"realy.lol/timestamp"
)

// GenerateChallenge creates a reasonable, 96 byte base64 challenge string
func GenerateChallenge() (b []byte) {
	bb := make([]byte, 12)
	b = make([]byte, 16)
	_, _ = rand.Read(bb)
	base64.StdEncoding.Encode(b, bb)
	return
}

// CreateUnsigned creates an event which should be sent via an "AUTH" command.
// If the authentication succeeds, the user will be authenticated as pubkey.
func CreateUnsigned(pubkey, challenge []byte, relayURL string) (ev *event.T) {
	return &event.T{
		PubKey:    pubkey,
		CreatedAt: timestamp.Now(),
		Kind:      kind.ClientAuthentication,
		Tags: tags.New(tag.New("relay", relayURL),
			tag.New("challenge", string(challenge))),
	}
}

// helper function for ValidateAuthEvent.
func parseURL(input string) (*url.URL, error) {
	return url.Parse(
		strings.ToLower(
			strings.TrimSuffix(input, "/"),
		),
	)
}

var ChallengeTag = []byte("challenge")
var RelayTag = []byte("relay")

// Validate checks whether event is a valid NIP-42 event for given challenge and relayURL.
// The result of the validation is encoded in the ok bool.
func Validate(evt *event.T, challenge []byte, relayURL string) (ok bool, err error) {
	// log.T.F("relayURL '%s'", relayURL)
	if evt.Kind.K != kind.ClientAuthentication.K {
		err = log.E.Err("event incorrect kind for auth: %d %s",
			evt.Kind.K, kind.GetString(evt.Kind))
		log.D.Ln(err)
		return
	}
	if evt.Tags.GetFirst(tag.New(ChallengeTag, challenge)) == nil {
		err = log.E.Err("challenge tag missing from auth response")
		log.D.Ln(err)
		return
	}
	// log.I.Ln(relayURL)
	var expected, found *url.URL
	if expected, err = parseURL(relayURL); chk.D(err) {
		log.D.Ln(err)
		return
	}
	r := evt.Tags.
		GetFirst(tag.New(RelayTag, nil)).Value()
	if len(r) == 0 {
		err = log.E.Err("relay tag missing from auth response")
		log.D.Ln(err)
		return
	}
	if found, err = parseURL(string(r)); chk.D(err) {
		err = log.E.Err("error parsing relay url: %s", err)
		log.D.Ln(err)
		return
	}
	if expected.Scheme != found.Scheme {
		err = log.E.Err("HTTP Scheme incorrect: expected '%s' got '%s",
			expected.Scheme, found.Scheme)
		log.D.Ln(err)
		return
	}
	if expected.Host != found.Host {
		err = log.E.Err("HTTP Host incorrect: expected '%s' got '%s",
			expected.Host, found.Host)
		log.D.Ln(err)
		return
	}
	if expected.Path != found.Path {
		err = log.E.Err("HTTP Path incorrect: expected '%s' got '%s",
			expected.Path, found.Path)
		log.D.Ln(err)
		return
	}

	now := time.Now()
	if evt.CreatedAt.Time().After(now.Add(10*time.Minute)) ||
		evt.CreatedAt.Time().Before(now.Add(-10*time.Minute)) {
		err = log.E.Err(
			"auth event more than 10 minutes before or after current time")
		log.D.Ln(err)
		return
	}
	// save for last, as it is the most expensive operation
	return evt.Verify()
}
