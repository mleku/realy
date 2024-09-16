package createdat

import (
	"bytes"
	"math"
	"testing"

	"lukechampine.com/frand"
	"mleku.dev/timestamp"
)

func TestT(t *testing.T) {
	for _ = range 1000000 {
		n := timestamp.FromUnix(int64(frand.Intn(math.MaxInt64)))
		v := New(n)
		buf := new(bytes.Buffer)
		v.Write(buf)
		buf2 := bytes.NewBuffer(buf.Bytes())
		v2 := New(timestamp.New())
		el := v2.Read(buf2).(*T)
		if el.Val.Int() != n.Int() {
			t.Fatalf("expected %d got %d", n.Int(), el.Val.Int())
		}
	}
}
