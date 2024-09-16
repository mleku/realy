package subscriptionid

import (
	"testing"

	"lukechampine.com/frand"
)

func TestMarshalJSONUnmarshalJSON(t *testing.T) {
	for _ = range 100 {
		b := make(B, frand.Intn(48)+1)
		bc := make(B, len(b))
		_, _ = frand.Read(b)
		copy(bc, b)
		var err error
		var si *T
		if si, err = New(b); chk.E(err) {
			t.Fatal(err)
		}
		var m B
		if m, err = si.MarshalJSON(nil); chk.E(err) {
			t.Fatal(err)
		}
		var ui *T
		ui, _ = New("")
		var rem B
		if rem, err = ui.UnmarshalJSON(m); chk.E(err) {
			t.Fatal(err)
		}
		if len(rem) > 0 {
			t.Errorf("len(rem): %d, '%s'", len(rem), rem)
		}
		if !equals(ui.T, bc) {
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
