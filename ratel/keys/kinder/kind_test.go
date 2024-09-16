package kinder

import (
	"bytes"
	"testing"

	"mleku.dev/kind"
)

func TestT(t *testing.T) {
	n := kind.New(1059)
	v := New(n.ToU16())
	buf := new(bytes.Buffer)
	v.Write(buf)
	buf2 := bytes.NewBuffer(buf.Bytes())
	v2 := New(0)
	el := v2.Read(buf2).(*T)
	if el.Val.ToU16() != n.ToU16() {
		t.Fatalf("expected %d got %d", n, el.Val)
	}
}
