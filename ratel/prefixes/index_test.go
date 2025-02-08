package prefixes

import (
	"bytes"
	"testing"

	"realy.lol/ratel/keys/index"
)

func TestT(t *testing.T) {
	v := Version.Key()
	// v := New(n)
	// buf := new(bytes.Buffer)
	// v.Write(buf)
	buf2 := bytes.NewBuffer(v)
	v2 := index.New(0)
	el := v2.Read(buf2).(*index.T)
	if el.Val[0] != v[0] {
		t.Fatalf("expected %d got %d", v[0], el.Val)
	}
}
