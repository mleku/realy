package ratel

import (
	"bytes"
	"runtime/debug"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dgraph-io/badger/v4"

	"github.com/clipperhouse/uax29/words"

	"realy.lol/chk"
	"realy.lol/event"
	"realy.lol/kind"
	"realy.lol/log"
	"realy.lol/ratel/keys/arb"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
)

type Words struct {
	ser     *serial.T
	wordMap map[string]struct{}
}

func (r *T) FulltextIndex() (err error) {
	r.WG.Add(1)
	defer r.WG.Done()
	wordsChan := make(chan Words)
	go func() {
		for {
			select {
			case <-r.Ctx.Done():
				return
			case w := <-wordsChan:
			retry:
				select {
				case <-r.Ctx.Done():
					return
				default:
				}
				for i := range w.wordMap {
					if err = r.Update(func(txn *badger.Txn) (err error) {
						prf := prefixes.FulltextIndex.Key(arb.New(i))
						var item2 *badger.Item
						if item2, err = txn.Get(prf); err != nil {
							// make a new record
							// log.I.F("making new index %s for %d", i, ser.Uint64())
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
						lprf := prefixes.FulltextLastIndexed.Key()
						if err = txn.Set(lprf, w.ser.Val); chk.E(err) {
							return
						}
						if w.ser.Uint64()%100 == 0 {
							debug.FreeOSMemory()
						}
						return
					}); chk.E(err) {
						time.Sleep(time.Second / 4)
						goto retry
					}
				}
			}
		}
	}()
	var last *serial.T
	if err = r.View(func(txn *badger.Txn) (err error) {
		var item *badger.Item
		if item, err = txn.Get(prefixes.FulltextLastIndexed.Key()); chk.E(err) {
			return
		}
		var val []byte
		if val, err = item.ValueCopy(nil); chk.E(err) {
			return
		}
		last = serial.New(val)
		return
	}); chk.E(err) {
	}
	if last == nil {
		last = serial.New(serial.Make(0))
	}
	if err = r.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: prefixes.Event.Key()})
		defer it.Close()
		for it.Seek(prefixes.Event.Key(last)); it.Valid(); it.Next() {
			item := it.Item()
			k := item.KeyCopy(nil)
			ser := serial.New(k[1:])
			if ser.Uint64() < last.Uint64() {
				k = k[:0]
				log.I.F("already done %d", ser.Uint64())
				continue
			}
			var val []byte
			if val, err = item.ValueCopy(nil); chk.E(err) {
				continue
			}
			ev := &event.T{}
			if _, err = r.Unmarshal(ev, val); chk.E(err) {
				return
			}
			wordMap := r.GetWordsFromContent(ev)
			if len(wordMap) > 0 {
				k, val = k[:0], val[:0]
				wordsChan <- Words{ser: ser, wordMap: wordMap}
				log.I.F("completed index %d", ser.Uint64())
			}
			wordMap = nil
			select {
			case <-r.Ctx.Done():
				log.I.F("context closed")
				return
			default:
			}
		}
		return
	}); chk.E(err) {
		return
	}
	return
}

func (r *T) GetWordsFromContent(ev *event.T) (wordMap map[string]struct{}) {
	wordMap = make(map[string]struct{})
	if ev.Kind.OneOf(kind.TextNote, kind.Article) {
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
				!bytes.HasPrefix(w, []byte("nostr:")) &&
				!bytes.HasPrefix(w, []byte("npub")) &&
				!bytes.HasPrefix(w, []byte("nsec")) &&
				!bytes.HasPrefix(w, []byte("nevent")) &&
				!bytes.HasPrefix(w, []byte("naddr")) &&
				!bytes.HasSuffix(w, []byte(".jpg")) &&
				!bytes.HasSuffix(w, []byte(".png")) &&
				!bytes.HasSuffix(w, []byte(".jpeg")) &&
				!bytes.HasSuffix(w, []byte(".mp4")) &&
				!bytes.HasSuffix(w, []byte(".mov")) &&
				!bytes.HasSuffix(w, []byte(".aac")) &&
				!bytes.HasSuffix(w, []byte(".mp3")) {
				wordMap[string(w)] = struct{}{}
			}
		}
		content = content[:0]
	}
	return
}
