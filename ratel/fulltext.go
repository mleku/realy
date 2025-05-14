package ratel

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/dgraph-io/badger/v4"

	"github.com/clipperhouse/uax29/words"

	"realy.lol/chk"
	"realy.lol/event"
	"realy.lol/hex"
	"realy.lol/ratel/keys/arb"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
)

type Words struct {
	ser     *serial.T
	wordMap map[string]struct{}
}

func (r *T) WriteFulltextIndex(w *Words) (err error) {
	if w == nil {
		return
	}
	r.WG.Add(1)
	defer r.WG.Done()
	for i := range w.wordMap {
	retry:
		if err = r.Update(func(txn *badger.Txn) (err error) {
			key := prefixes.FulltextIndex.Key(arb.New(i), w.ser)
			if err = txn.Set(key, nil); chk.E(err) {
				return
			}
			return
		}); chk.E(err) {
			goto retry
		}
	}
	return
}

func (r *T) GetFulltextKeys(ev *event.T, ser *serial.T) (keys [][]byte) {
	w := r.GetWordsFromContent(ev)
	for i := range w {
		key := prefixes.FulltextIndex.Key(arb.New(i), ser)
		keys = append(keys, key)
	}
	return
}

func (r *T) GetWordsFromContent(ev *event.T) (wordMap map[string]struct{}) {
	wordMap = make(map[string]struct{})
	if ev.Kind.IsText() {
		content := ev.Content
		seg := words.NewSegmenter(content)
		for seg.Next() {
			w := seg.Bytes()
			w = bytes.ToLower(w)
			var ru rune
			ru, _ = utf8.DecodeRune(w)
			// ignore the most common things that aren't words\
			if !unicode.IsSpace(ru) &&
				!unicode.IsPunct(ru) &&
				!unicode.IsSymbol(ru) &&
				!bytes.HasSuffix(w, []byte(".jpg")) &&
				!bytes.HasSuffix(w, []byte(".png")) &&
				!bytes.HasSuffix(w, []byte(".jpeg")) &&
				!bytes.HasSuffix(w, []byte(".mp4")) &&
				!bytes.HasSuffix(w, []byte(".mov")) &&
				!bytes.HasSuffix(w, []byte(".aac")) &&
				!bytes.HasSuffix(w, []byte(".mp3")) &&
				!IsEntity(w) &&
				!bytes.Contains(w, []byte(".")) {
				if len(w) == 64 || len(w) == 128 {
					if _, err := hex.Dec(string(w)); !chk.E(err) {
						continue
					}
				}

				wordMap[string(w)] = struct{}{}
			}
		}
		content = content[:0]
	}
	return
}

func IsEntity(w []byte) (is bool) {
	var b []byte
	b = []byte("nostr:")
	if bytes.Contains(w, b) && len(b) < len(w) {
		return true
	}
	b = []byte("npub")
	if bytes.Contains(w, b) && len(b) < len(w) {
		return true
	}
	b = []byte("nsec")
	if bytes.Contains(w, b) && len(b) < len(w) {
		return true
	}
	b = []byte("nevent")
	if bytes.Contains(w, b) && len(b) < len(w) {
		return true
	}
	b = []byte("naddr")
	if bytes.Contains(w, b) && len(b) < len(w) {
		return true
	}
	b = []byte("cashu")
	if bytes.Contains(w, b) && len(b) < len(w) {
		return true
	}
	return
}
