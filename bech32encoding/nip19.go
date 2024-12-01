package bech32encoding

import (
	"bytes"
	"encoding/binary"
	"reflect"

	"realy.lol/bech32encoding/pointers"
	"realy.lol/ec/bech32"
	"realy.lol/ec/schnorr"
	"realy.lol/eventid"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/sha256"
)

var (
	NoteHRP     = by("note")
	NsecHRP     = by("nsec")
	NpubHRP     = by("npub")
	NprofileHRP = by("nprofile")
	NeventHRP   = by("nevent")
	NentityHRP  = by("naddr")
)

func DecodeToString(bech32String by) (prefix, value by, err er) {
	var s any
	if prefix, s, err = Decode(bech32String); chk.D(err) {
		return
	}
	var ok bo
	if value, ok = s.(by); ok {
		return
	}
	err = log.E.Err("value was not decoded to a string, found type %s",
		reflect.TypeOf(s))
	return
}

func Decode(bech32string by) (prefix by, value any, err er) {
	var bits5 by
	if prefix, bits5, err = bech32.DecodeNoLimit(bech32string); chk.D(err) {
		return
	}
	var data []byte
	if data, err = bech32.ConvertBits(bits5, 5, 8, false); chk.D(err) {
		return prefix, nil, errorf.E("failed translating data into 8 bits: %s", err.Error())
	}
	switch {
	case equals(prefix, NpubHRP) ||
		equals(prefix, NsecHRP) ||
		equals(prefix, NoteHRP):
		if len(data) < 32 {
			return prefix, nil, errorf.E("data is less than 32 bytes (%d)", len(data))
		}
		b := make(by, schnorr.PubKeyBytesLen*2)
		hex.EncBytes(b, data[:32])
		return prefix, b, nil
	case equals(prefix, NprofileHRP):
		var result pointers.Profile
		curr := 0
		for {
			t, v := readTLVEntry(data[curr:])
			if v == nil {
				// end here
				if len(result.PublicKey) < 1 {
					return prefix, result, errorf.E("no pubkey found for nprofile")
				}
				return prefix, result, nil
			}
			switch t {
			case TLVDefault:
				if len(v) < 32 {
					return prefix, nil, errorf.E("pubkey is less than 32 bytes (%d)", len(v))
				}
				result.PublicKey = make(by, schnorr.PubKeyBytesLen*2)
				hex.EncBytes(result.PublicKey, v)
			case TLVRelay:
				result.Relays = append(result.Relays, v)
			default:
				// ignore
			}
			curr = curr + 2 + len(v)
		}
	case equals(prefix, NeventHRP):
		var result pointers.Event
		curr := 0
		for {
			t, v := readTLVEntry(data[curr:])
			if v == nil {
				// end here
				if result.ID.Len() == 0 {
					return prefix, result, errorf.E("no id found for nevent")
				}
				return prefix, result, nil
			}
			switch t {
			case TLVDefault:
				if len(v) < 32 {
					return prefix, nil, errorf.E("id is less than 32 bytes (%d)", len(v))
				}
				result.ID, err = eventid.NewFromBytes(v)
			case TLVRelay:
				result.Relays = append(result.Relays, v)
			case TLVAuthor:
				if len(v) < 32 {
					return prefix, nil, errorf.E("author is less than 32 bytes (%d)", len(v))
				}
				result.Author = make(by, schnorr.PubKeyBytesLen*2)
				hex.EncBytes(result.Author, v)
			case TLVKind:
				result.Kind = kind.New(binary.BigEndian.Uint32(v))
			default:
				// ignore
			}
			curr = curr + 2 + len(v)
		}
	case equals(prefix, NentityHRP):
		var result pointers.Entity
		curr := 0
		for {
			t, v := readTLVEntry(data[curr:])
			if v == nil {
				// end here
				if result.Kind.ToU16() == 0 ||
					len(result.Identifier) < 1 ||
					len(result.PublicKey) < 1 {

					return prefix, result, errorf.E("incomplete naddr")
				}
				return prefix, result, nil
			}
			switch t {
			case TLVDefault:
				result.Identifier = v
			case TLVRelay:
				result.Relays = append(result.Relays, v)
			case TLVAuthor:
				if len(v) < 32 {
					return prefix, nil, errorf.E("author is less than 32 bytes (%d)", len(v))
				}
				result.PublicKey = make(by, schnorr.PubKeyBytesLen*2)
				hex.EncBytes(result.PublicKey, v)
			case TLVKind:
				result.Kind = kind.New(binary.BigEndian.Uint32(v))
			default:
				log.D.Ln("got a bogus TLV type code", t)
				// ignore
			}
			curr = curr + 2 + len(v)
		}
	}
	return prefix, data, errorf.E("unknown tag %s", prefix)
}

func EncodeNote(eventIDHex by) (s by, err er) {
	var b []byte
	if _, err = hex.DecBytes(b, eventIDHex); chk.D(err) {
		err = log.E.Err("failed to decode event id hex: %w", err)
		return
	}
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(b, 8, 5, true); chk.D(err) {
		return
	}
	return bech32.Encode(NoteHRP, bits5)
}

func EncodeProfile(publicKeyHex by, relays []by) (s by, err er) {
	buf := &bytes.Buffer{}
	pb := make(by, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(pb, publicKeyHex); chk.D(err) {
		err = log.E.Err("invalid pubkey '%s': %w", publicKeyHex, err)
		return
	}
	writeTLVEntry(buf, TLVDefault, pb)
	for _, url := range relays {
		writeTLVEntry(buf, TLVRelay, []byte(url))
	}
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(buf.Bytes(), 8, 5, true); chk.D(err) {
		err = log.E.Err("failed to convert bits: %w", err)
		return
	}
	return bech32.Encode(NprofileHRP, bits5)
}

func EncodeEvent(eventIDHex *eventid.T, relays []by, author by) (s by, err er) {
	buf := &bytes.Buffer{}
	id := make(by, sha256.Size)
	if _, err = hex.DecBytes(id, eventIDHex.ByteString(nil)); chk.D(err) ||
		len(id) != 32 {

		return nil, errorf.E("invalid id %d '%s': %v", len(id), eventIDHex,
			err)
	}
	writeTLVEntry(buf, TLVDefault, id)
	for _, url := range relays {
		writeTLVEntry(buf, TLVRelay, []byte(url))
	}
	pubkey := make(by, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(pubkey, author); len(pubkey) == 32 {
		writeTLVEntry(buf, TLVAuthor, pubkey)
	}
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(buf.Bytes(), 8, 5, true); chk.D(err) {
		err = log.E.Err("failed to convert bits: %w", err)
		return
	}

	return bech32.Encode(NeventHRP, bits5)
}

func EncodeEntity(pk by, k *kind.T, id by, relays []by) (s by, err er) {
	buf := &bytes.Buffer{}
	writeTLVEntry(buf, TLVDefault, []byte(id))
	for _, url := range relays {
		writeTLVEntry(buf, TLVRelay, []byte(url))
	}
	pb := make(by, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(pb, pk); chk.D(err) {
		return nil, errorf.E("invalid pubkey '%s': %w", pb, err)
	}
	writeTLVEntry(buf, TLVAuthor, pb)
	kindBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(kindBytes, uint32(k.ToU16()))
	writeTLVEntry(buf, TLVKind, kindBytes)
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(buf.Bytes(), 8, 5, true); chk.D(err) {
		return nil, errorf.E("failed to convert bits: %w", err)
	}
	return bech32.Encode(NentityHRP, bits5)
}
