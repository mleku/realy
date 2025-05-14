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
	// log.I.F("making fulltext index for %d", w.ser.Uint64())
	for i := range w.wordMap {
	retry:
		if err = r.Update(func(txn *badger.Txn) (err error) {
			prf := prefixes.FulltextIndex.Key(arb.New(i))
			var item2 *badger.Item
			if item2, err = txn.Get(prf); err != nil {
				// make a new record
				if err = txn.Set(prf, w.ser.Val); chk.E(err) {
					return
				}
			} else {
				if item2.KeySize() == int64(len(prf)) {
					select {
					case <-r.Ctx.Done():
						return
					default:
					}
					var val2 []byte
					if val2, err = item2.ValueCopy(nil); chk.E(err) {
						return
					}
					if !bytes.Contains(val2, w.ser.Val) {
						val2 = append(val2, w.ser.Val...)
						if err = txn.Set(prf, val2); chk.E(err) {
							return
						}
					}
					return
				}
			}
			return
		}); chk.E(err) {
			goto retry
		}
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
