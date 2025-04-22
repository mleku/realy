package text

import (
	"testing"

	"lukechampine.com/frand"

	"realy.mleku.dev/chk"
	"realy.mleku.dev/sha256"
)

func TestUnescapeByteString(t *testing.T) {
	b := make([]byte, 256)
	for i := range b {
		b[i] = byte(i)
	}
	escaped := NostrEscape(nil, b)
	unescaped := NostrUnescape(escaped)
	if string(b) != string(unescaped) {
		t.Log(b)
		t.Log(unescaped)
		t.FailNow()
	}
}

func GenRandString(l int, src *frand.RNG) (str []byte) {
	return src.Bytes(l)
}

var seed = sha256.Sum256([]byte(`
The tao that can be told
is not the eternal Tao
The name that can be named
is not the eternal Name

The unnamable is the eternally real
Naming is the origin of all particular things

Free from desire, you realize the mystery
Caught in desire, you see only the manifestations

Yet mystery and manifestations arise from the same source
This source is called darkness

Darkness within darkness
The gateway to all understanding
`))

var src = frand.NewCustom(seed[:], 32, 12)

func TestRandomEscapeByteString(t *testing.T) {
	// this is a kind of fuzz test, does a massive number of iterations of
	// random content that ensures the escaping is correct without creating a
	// fixed set of test vectors.

	for i := 0; i < 1000; i++ {
		l := src.Intn(1<<8) + 32
		s1 := GenRandString(l, src)
		s2 := make([]byte, l)
		orig := make([]byte, l)
		copy(s2, s1)
		copy(orig, s1)

		// first we are checking our implementation comports to the one from go-nostr.
		escapeStringVersion := NostrEscape([]byte{}, s1)
		escapeJSONStringAndWrapVersion := NostrEscape(nil, s2)
		if len(escapeJSONStringAndWrapVersion) != len(escapeStringVersion) {
			t.Logf("escapeString\nlength: %d\n%s\n%v\n",
				len(escapeStringVersion), string(escapeStringVersion),
				escapeStringVersion)
			t.Logf("escapJSONStringAndWrap\nlength: %d\n%s\n%v\n",
				len(escapeJSONStringAndWrapVersion),
				escapeJSONStringAndWrapVersion,
				escapeJSONStringAndWrapVersion)
			t.FailNow()
		}
		for i := range escapeStringVersion {
			if i > len(escapeJSONStringAndWrapVersion) {
				t.Fatal("escapeString version is shorter")
			}
			if escapeStringVersion[i] != escapeJSONStringAndWrapVersion[i] {
				t.Logf("escapeString version differs at index %d from "+
					"escapeJSONStringAndWrap version\n%s\n%s\n%v\n%v", i,
					escapeStringVersion[i-4:],
					escapeJSONStringAndWrapVersion[i-4:],
					escapeStringVersion[i-4:],
					escapeJSONStringAndWrapVersion[i-4:])
				t.Logf("escapeString\nlength: %d %s\n",
					len(escapeStringVersion), escapeStringVersion)
				t.Logf("escapJSONStringAndWrap\nlength: %d %s\n",
					len(escapeJSONStringAndWrapVersion),
					escapeJSONStringAndWrapVersion)
				t.Logf("got '%s' %d expected '%s' %d\n",
					string(escapeJSONStringAndWrapVersion[i]),
					escapeJSONStringAndWrapVersion[i],
					string(escapeStringVersion[i]),
					escapeStringVersion[i],
				)
				t.FailNow()
			}
		}

		// next, unescape the output and see if it matches the original
		unescaped := NostrUnescape(escapeJSONStringAndWrapVersion)
		// t.Logf("unescaped: \n%s\noriginal:  \n%s", unescaped, orig)
		if string(unescaped) != string(orig) {
			t.Fatalf("\ngot      %d %v\nexpected %d %v\n",
				len(unescaped),
				unescaped,
				len(orig),
				orig,
			)
		}
	}
}

func BenchmarkNostrEscapeNostrUnescape(b *testing.B) {
	const size = 65536
	b.Run("frand64k", func(b *testing.B) {
		b.ReportAllocs()
		in := make([]byte, size)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
		}
	})
	b.Run("NostrEscape64k", func(b *testing.B) {
		b.ReportAllocs()
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			out = out[:0]
		}
	})
	b.Run("NostrEscapeNostrUnescape64k", func(b *testing.B) {
		b.ReportAllocs()
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			in = in[:0]
			out = NostrUnescape(out)
			out = out[:0]
		}
	})
	b.Run("frand32k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 2
		in := make([]byte, size)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
		}
	})
	b.Run("NostrEscape32k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 2
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			out = out[:0]
		}
	})
	b.Run("NostrEscapeNostrUnescape32k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 2
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			in = in[:0]
			out = NostrUnescape(out)
			out = out[:0]
		}
	})
	b.Run("frand16k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 4
		in := make([]byte, size)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
		}
	})
	b.Run("NostrEscape16k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 4
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			out = out[:0]
		}
	})
	b.Run("NostrEscapeNostrUnescape16k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 4
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			in = in[:0]
			out = NostrUnescape(out)
			out = out[:0]
		}
	})
	b.Run("frand8k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 8
		in := make([]byte, size)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
		}
	})
	b.Run("NostrEscape8k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 8
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			out = out[:0]
		}
	})
	b.Run("NostrEscapeNostrUnescape8k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 8
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			in = in[:0]
			out = NostrUnescape(out)
			out = out[:0]
		}
	})
	b.Run("frand4k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 16
		in := make([]byte, size)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
		}
	})
	b.Run("NostrEscape4k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 16
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			out = out[:0]
		}
	})
	b.Run("NostrEscapeNostrUnescape4k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 16
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			in = in[:0]
			out = NostrUnescape(out)
			out = out[:0]
		}
	})
	b.Run("frand2k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 32
		in := make([]byte, size)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
		}
	})
	b.Run("NostrEscape2k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 32
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			out = out[:0]
		}
	})
	b.Run("NostrEscapeNostrUnescape2k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 32
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			in = in[:0]
			out = NostrUnescape(out)
			out = out[:0]
		}
	})
	b.Run("frand1k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 64
		in := make([]byte, size)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
		}
	})
	b.Run("NostrEscape1k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 64
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			out = out[:0]
		}
	})
	b.Run("NostrEscapeNostrUnescape1k", func(b *testing.B) {
		b.ReportAllocs()
		size := size / 64
		in := make([]byte, size)
		out := make([]byte, size*2)
		var err error
		for i := 0; i < b.N; i++ {
			if _, err = frand.Read(in); chk.E(err) {
				b.Fatal(err)
			}
			out = NostrEscape(out, in)
			in = in[:0]
			out = NostrUnescape(out)
			out = out[:0]
		}
	})
}
