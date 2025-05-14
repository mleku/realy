package ratel

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/dgraph-io/badger/v4"

	"github.com/clipperhouse/uax29/words"

	"realy.lol/chk"
	"realy.lol/event"
	"realy.lol/eventid"
	"realy.lol/hex"
	"realy.lol/ratel/keys/arb"
	"realy.lol/ratel/keys/createdat"
	"realy.lol/ratel/keys/fullid"
	"realy.lol/ratel/keys/integer"
	"realy.lol/ratel/keys/kinder"
	"realy.lol/ratel/keys/pubkey"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
)

type Words struct {
	ser     *serial.T
	ev      *event.T
	wordMap map[string]int
}

func (r *T) WriteFulltextIndex(w *Words) (err error) {
	if w == nil {
		return
	}
	r.WG.Add(1)
	defer r.WG.Done()
	for word, pos := range w.wordMap {
	retry:
		if err = r.Update(func(txn *badger.Txn) (err error) {
			var eid *eventid.T
			if eid, err = eventid.NewFromBytes(w.ev.Id); chk.E(err) {
				return
			}
			var pk *pubkey.T
			if pk, err = pubkey.New(w.ev.Pubkey); chk.E(err) {
				return
			}
			key := prefixes.FulltextIndex.Key(
				arb.New(word),
				fullid.New(eid),
				pk,
				createdat.New(w.ev.CreatedAt),
				kinder.New(w.ev.Kind.ToU16()),
				integer.New(pos),
				w.ser,
			)
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

func (r *T) GetWordsFromContent(ev *event.T) (wordMap map[string]int) {
	wordMap = make(map[string]int)
	if ev.Kind.IsText() {
		content := ev.Content
		seg := words.NewSegmenter(content)
		var counter int
		for seg.Next() {
			w := seg.Bytes()
			w = bytes.ToLower(w)
			var ru rune
			ru, _ = utf8.DecodeRune(w)
			// ignore the most common things that aren't words
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
				wordMap[string(w)] = counter
				counter++
			}
		}
		content = content[:0]
	}
	return
}

func IsEntity(w []byte) (is bool) {
	var b []byte
	b = []byte("nostr:")
	if bytes.Contains(w, b) && len(b)+10 < len(w) {
		return true
	}
	b = []byte("npub")
	if bytes.Contains(w, b) && len(b)+5 < len(w) {
		return true
	}
	b = []byte("nsec")
	if bytes.Contains(w, b) && len(b)+5 < len(w) {
		return true
	}
	b = []byte("nevent")
	if bytes.Contains(w, b) && len(b)+5 < len(w) {
		return true
	}
	b = []byte("naddr")
	if bytes.Contains(w, b) && len(b)+5 < len(w) {
		return true
	}
	b = []byte("note")
	if bytes.Contains(w, b) && len(b)+20 < len(w) {
		return true
	}
	b = []byte("lnurl")
	if bytes.Contains(w, b) && len(b)+20 < len(w) {
		return true
	}
	b = []byte("cashu")
	if bytes.Contains(w, b) && len(b)+20 < len(w) {
		return true
	}
	return
}
