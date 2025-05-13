package ratel

import (
	"time"

	"github.com/dgraph-io/badger/v4"

	"realy.lol/chk"
	"realy.lol/event"
	"realy.lol/kind"
	"realy.lol/log"
	"realy.lol/ratel/keys/lang"
	"realy.lol/ratel/keys/serial"
	"realy.lol/ratel/prefixes"
	"realy.lol/tag"
)

type Langs struct {
	ser   *serial.T
	langs []string
}

// LangIndex searches through events for language tags and stores a LangIndex key containing the
// ISO-639-2 language code and serial to search for text events by language.
func (r *T) LangIndex() (err error) {
	r.WG.Add(1)
	defer r.WG.Done()
	log.I.F("indexing language tags")
	defer log.I.F("finished indexing language tags")
	langChan := make(chan Langs)
	go func() {
		for {
			select {
			case <-r.Ctx.Done():
				return
			case l := <-langChan:
				if len(l.langs) < 1 {
					continue
				}
				log.I.S("making lang index for %d %v", l.ser.Uint64(), l.langs)
			retry:
				if err = r.Update(func(txn *badger.Txn) (err error) {
					for _, v := range l.langs {
						select {
						case <-r.Ctx.Done():
							return
						default:
						}
						key := prefixes.LangIndex.Key(lang.New(v), l.ser)
						if err = txn.Set(key, nil); chk.E(err) {
							return
						}
						return
					}
					return
				}); chk.E(err) {
					time.Sleep(time.Second / 4)
					goto retry
				}

			}
		}
	}()
	var last *serial.T
	if err = r.View(func(txn *badger.Txn) (err error) {
		var item *badger.Item
		if item, err = txn.Get(prefixes.LangLastIndexed.Key()); chk.E(err) {
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
	if err = r.Update(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.IteratorOptions{Prefix: prefixes.Event.Key()})
		defer it.Close()
		for it.Seek(prefixes.Event.Key(last)); it.Valid(); it.Next() {
			item := it.Item()
			k := item.KeyCopy(nil)
			ser := serial.New(k[1:])
			log.I.F("lang index scanning %d", ser.Uint64())
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
			langs := r.GetLangTags(ev)
			lprf := prefixes.LangLastIndexed.Key()
			if err = txn.Set(lprf, ser.Val); chk.E(err) {
				return
			}
			if len(langs) > 0 {
				l := Langs{ser: ser, langs: langs}
				log.I.S(l)
				langChan <- l
			}
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

func (r *T) GetLangTags(ev *event.T) (langs []string) {
	if ev.Kind.OneOf(kind.TextNote, kind.Article) {
		tgs := ev.Tags.GetAll(tag.New("l"))
		tgsl := tgs.ToStringsSlice()
		for _, v := range tgsl {
			for _, w := range LanguageCodes {
				if v[1] == w.ISO639_1 || v[1] == w.ISO639_2 {
					langs = append(langs, w.ISO639_2)
				}
			}
		}
	}
	return
}
