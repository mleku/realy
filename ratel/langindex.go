package ratel

import (
	"github.com/dgraph-io/badger/v4"

	"realy.lol/chk"
	"realy.lol/event"
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

func (r *T) WriteLangIndex(l *Langs) (err error) {
	if len(l.langs) > 0 {
		log.I.F("making lang index for %d", l.ser.Uint64())
	} else {
		return
	}
	r.WG.Add(1)
	defer r.WG.Done()
	log.I.F("making lang index for %d", l.ser.Uint64())
retry:
	if err = r.Update(func(txn *badger.Txn) (err error) {
		for _, v := range l.langs {
			log.I.F("lang %s on %d", v, l.ser.Uint64())
			select {
			case <-r.Ctx.Done():
				return
			default:
			}
			key := prefixes.LangIndex.Key(lang.New(v), l.ser)
			if err = txn.Set(key, nil); chk.E(err) {
				return
			}
			log.I.F("wrote lang index for %d", l.ser.Uint64())
			return
		}
		return
	}); chk.E(err) {
		goto retry
	}
	return
}

func (r *T) GetLangKeys(ev *event.T, ser *serial.T) (keys [][]byte) {
	langs := r.GetLangTags(ev)
	for _, v := range langs {
		key := prefixes.LangIndex.Key(lang.New(v), ser)
		keys = append(keys, key)
	}
	return
}

func (r *T) GetLangTags(ev *event.T) (langs []string) {
	if ev.Kind.IsText() {
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
