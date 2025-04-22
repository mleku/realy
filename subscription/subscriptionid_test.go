package subscription

import (
	"bytes"
	"testing"

	"lukechampine.com/frand"

	"realy.mleku.dev/chk"
)

func TestMarshalUnmarshal(t *testing.T) {
	for _ = range 100 {
		b := make([]byte, frand.Intn(48)+1)
		bc := make([]byte, len(b))
		_, _ = frand.Read(b)
		copy(bc, b)
		var err error
		var si *Id
		if si, err = NewId(b); chk.E(err) {
			t.Fatal(err)
		}
		var m []byte
		m = si.Marshal(nil)
		var ui *Id
		ui, _ = NewId("")
		var rem []byte
		if rem, err = ui.Unmarshal(m); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) > 0 {
			t.Errorf("len(rem): %d, '%s'", len(rem), rem)
		}
		if !bytes.Equal(ui.T, bc) {
			t.Fatalf("bc: %0x, uu: %0x", bc, ui)
		}
	}
}

func TestNewStd(t *testing.T) {
	for _ = range 100 {
		if NewStd() == nil {
			t.Fatal("NewStd() returned nil")
		}
	}
}
