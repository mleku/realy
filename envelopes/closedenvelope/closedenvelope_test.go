package closedenvelope

import (
	"testing"

	"lukechampine.com/frand"
	"mleku.dev/envelopes"
	"mleku.dev/subscriptionid"
)

var messages = []B{
	B(""),
	B("pow: difficulty 25>=24"),
	B("duplicate: already have this event"),
	B("blocked: you are banned from posting here"),
	B("blocked: please register your pubkey at https://my-expensive-relay.example.com"),
	B("rate-limited: slow down there chief"),
	B("invalid: event creation date is too far off from the current time"),
	B("pow: difficulty 26 is less than 30"),
	B("error: could not connect to the database"),
}

func RandomMessage() B {
	return messages[frand.Intn(len(messages)-1)]
}

func TestMarshalJSONUnmarshalJSON(t *testing.T) {
	var err error
	rb, rb1, rb2 := make(B, 0, 65535), make(B, 0, 65535), make(B, 0, 65535)
	for _ = range 1000 {
		var s *subscriptionid.T
		if s = subscriptionid.NewStd(); chk.E(err) {
			t.Fatal(err)
		}
		req := NewFrom(s, RandomMessage())
		if rb, err = req.MarshalJSON(rb); chk.E(err) {
			t.Fatal(err)
		}
		rb1 = rb1[:len(rb)]
		copy(rb1, rb)
		var rem B
		var l string
		if l, rb, err = envelopes.Identify(rb); chk.E(err) {
			t.Fatal(err)
		}
		if l != L {
			t.Fatalf("invalid sentinel %s, expect %s", l, L)
		}
		req2 := New()
		if rem, err = req2.UnmarshalJSON(rb); chk.E(err) {
			t.Fatal(err)
		}
		// log.I.Ln(req2.ID)
		if len(rem) > 0 {
			t.Fatalf("unmarshal failed, remainder\n%d %s",
				len(rem), rem)
		}
		if rb2, err = req2.MarshalJSON(rb2); chk.E(err) {
			t.Fatal(err)
		}
		if !equals(rb1, rb2) {
			if len(rb1) != len(rb2) {
				t.Fatalf("unmarshal failed, different lengths\n%d %s\n%d %s\n",
					len(rb1), rb1, len(rb2), rb2)
			}
			for i := range rb1 {
				if rb1[i] != rb2[i] {
					t.Fatalf("unmarshal failed, difference at position %d\n%d %s\n%s\n%d %s\n%s\n",
						i, len(rb1), rb1[:i], rb1[i:], len(rb2), rb2[:i],
						rb2[i:])
				}
			}
			t.Fatalf("unmarshal failed\n%d %s\n%d %s\n",
				len(rb1), rb1, len(rb2), rb2)
		}
		rb, rb1, rb2 = rb[:0], rb1[:0], rb2[:0]
	}
}