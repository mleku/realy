package tombstone

import (
	"bytes"
	"testing"

	"lukechampine.com/frand"

	"realy.mleku.dev/eventid"
)

func TestT(t *testing.T) {
	id := frand.Entropy256()
	ts := NewWith(eventid.NewWith(id[:]))
	buf := new(bytes.Buffer)
	ts.Write(buf)
	buf2 := bytes.NewBuffer(buf.Bytes())
	ts2 := New()
	ts2.Read(buf2)
	if !bytes.Equal(ts.val, ts2.val) {
		t.Errorf("expected %0x got %0x", ts.val, ts2.val)
	}
}
