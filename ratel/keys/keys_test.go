// package keys_test needs to be a different package name or the implementation
// types imports will circular
package keys_test

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"lukechampine.com/frand"

	"realy.mleku.dev/ec/schnorr"
	"realy.mleku.dev/eventid"
	"realy.mleku.dev/kind"
	"realy.mleku.dev/ratel/keys"
	"realy.mleku.dev/ratel/keys/createdat"
	"realy.mleku.dev/ratel/keys/id"
	"realy.mleku.dev/ratel/keys/index"
	"realy.mleku.dev/ratel/keys/kinder"
	"realy.mleku.dev/ratel/keys/pubkey"
	"realy.mleku.dev/ratel/keys/serial"
	"realy.mleku.dev/ratel/prefixes"
	"realy.mleku.dev/timestamp"
)

func TestElement(t *testing.T) {
	for _ = range 100000 {
		var failed bool
		{ // construct a typical key type of structure
			// a prefix
			np := prefixes.Version
			vp := index.New(byte(np))
			// an id
			fakeIdBytes := frand.Bytes(sha256.Size)
			i := eventid.NewWith(fakeIdBytes)
			vid := id.New(i)
			// a kinder
			n := kind.New(1059)
			vk := kinder.New(n.K)
			// a pubkey
			fakePubkeyBytes := frand.Bytes(schnorr.PubKeyBytesLen)
			var vpk *pubkey.T
			var err error
			vpk, err = pubkey.NewFromBytes(fakePubkeyBytes)
			if err != nil {
				t.Fatal(err)
			}
			// a createdat
			ts := timestamp.Now()
			vca := createdat.New(ts)
			// a serial
			fakeSerialBytes := frand.Bytes(serial.Len)
			vs := serial.New(fakeSerialBytes)
			// write Element list into buffer
			b := keys.Write(vp, vid, vk, vpk, vca, vs)
			// check that values decoded all correctly
			// we expect the following types, so we must create them:
			var vp2 = index.New(0)
			var vid2 = id.New()
			var vk2 = kinder.New(0)
			var vpk2 *pubkey.T
			vpk2, err = pubkey.New()
			if err != nil {
				t.Fatal(err)
			}
			var vca2 = createdat.New(timestamp.New())
			var vs2 = serial.New(nil)
			// read it in
			keys.Read(b, vp2, vid2, vk2, vpk2, vca2, vs2)
			// this is a lot of tests, so use switch syntax
			switch {
			case bytes.Compare(vp.Val, vp2.Val) != 0:
				t.Logf("failed to decode correctly got %v expected %v", vp2.Val,
					vp.Val)
				failed = true
				fallthrough
			case bytes.Compare(vid.Val, vid2.Val) != 0:
				t.Logf("failed to decode correctly got %v expected %v", vid2.Val,
					vid.Val)
				failed = true
				fallthrough
			case vk.Val.ToU16() != vk2.Val.ToU16():
				t.Logf("failed to decode correctly got %v expected %v", vk2.Val,
					vk.Val)
				failed = true
				fallthrough
			case !bytes.Equal(vpk.Val, vpk2.Val):
				t.Logf("failed to decode correctly got %v expected %v", vpk2.Val,
					vpk.Val)
				failed = true
				fallthrough
			case vca.Val.I64() != vca2.Val.I64():
				t.Logf("failed to decode correctly got %v expected %v", vca2.Val,
					vca.Val)
				failed = true
				fallthrough
			case !bytes.Equal(vs.Val, vs2.Val):
				t.Logf("failed to decode correctly got %v expected %v", vpk2.Val,
					vpk.Val)
				failed = true
			}
		}
		{ // construct a counter value
			// a createdat
			ts := timestamp.Now()
			vca := createdat.New(ts)
			// a sizer
			// n := uint32(frand.Uint64n(math.MaxUint32))
			// write out values
			b := keys.Write(vca)
			// check that values decoded all correctly
			// we expect the following types, so we must create them:
			var vca2 = createdat.New(timestamp.New())
			// read it in
			keys.Read(b, vca2)
			// check they match

			if vca.Val.I64() != vca2.Val.I64() {
				t.Logf("failed to decode correctly got %v expected %v", vca2.Val,
					vca.Val)
				failed = true
			}
		}
		if failed {
			t.FailNow()
		}
	}
}
