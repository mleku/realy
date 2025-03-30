package bech32encoding

import (
	"bytes"
	"encoding/binary"

	"realy.lol/bech32encoding/pointers"
	"realy.lol/bech32encoding/tlv"
	"realy.lol/ec/bech32"
	"realy.lol/ec/schnorr"
	"realy.lol/eventid"
	"realy.lol/hex"
	"realy.lol/kind"
	"realy.lol/sha256"
)

var (
	// NoteHRP is the Human Readable Prefix (HRP) for a nostr note (kind 1)
	NoteHRP = []byte("note")

	// NsecHRP is the Human Readable Prefix (HRP) for a nostr secret key
	NsecHRP = []byte("nsec")

	// NpubHRP is the Human Readable Prefix (HRP) for a nostr public key
	NpubHRP = []byte("npub")

	// NprofileHRP is the Human Readable Prefix (HRP) for a nostr profile metadata
	// event (kind 0)
	NprofileHRP = []byte("nprofile")

	// NeventHRP is the Human Readable Prefix (HRP) for a nostr event, which may
	// include relay hints to find the event, and the author's npub.
	NeventHRP = []byte("nevent")

	// NentityHRP is the Human Readable Prefix (HRP) for a nostr is a generic nostr entity, which may include relay hints to find the event, and the author's npub.
	NentityHRP = []byte("naddr")
)

// Decode a nostr bech32 encoded entity, return the prefix, and the decoded
// value, and any error if one occurred in the process of decoding.
func Decode(bech32string []byte) (prefix []byte, value any, err error) {
	var bits5 []byte
	if prefix, bits5, err = bech32.DecodeNoLimit(bech32string); chk.D(err) {
		return
	}
	var data []byte
	if data, err = bech32.ConvertBits(bits5, 5, 8, false); chk.D(err) {
		return prefix, nil, errorf.E("failed translating data into 8 bits: %s", err.Error())
	}
	buf := bytes.NewBuffer(data)
	switch {
	case bytes.Equal(prefix, NpubHRP) ||
		bytes.Equal(prefix, NsecHRP) ||
		bytes.Equal(prefix, NoteHRP):
		if len(data) < 32 {
			return prefix, nil, errorf.E("data is less than 32 bytes (%d)", len(data))
		}
		b := make([]byte, schnorr.PubKeyBytesLen*2)
		hex.EncBytes(b, data[:32])
		return prefix, b, nil
	case bytes.Equal(prefix, NprofileHRP):
		var result pointers.Profile
		for {
			t, v := tlv.ReadEntry(buf)
			if len(v) == 0 {
				// end here
				if len(result.PublicKey) < 1 {
					return prefix, result, errorf.E("no pubkey found for nprofile")
				}
				return prefix, result, nil
			}
			switch t {
			case tlv.Default:
				if len(v) < 32 {
					return prefix, nil, errorf.E("pubkey is less than 32 bytes (%d)", len(v))
				}
				result.PublicKey = make([]byte, schnorr.PubKeyBytesLen*2)
				hex.EncBytes(result.PublicKey, v)
			case tlv.Relay:
				result.Relays = append(result.Relays, v)
			default:
				// ignore
			}
		}
	case bytes.Equal(prefix, NeventHRP):
		var result pointers.Event
		for {
			t, v := tlv.ReadEntry(buf)
			if v == nil {
				// end here
				if result.ID.Len() == 0 {
					return prefix, result, errorf.E("no id found for nevent")
				}
				return prefix, result, nil
			}
			switch t {
			case tlv.Default:
				if len(v) < 32 {
					return prefix, nil, errorf.E("id is less than 32 bytes (%d)", len(v))
				}
				result.ID = eventid.NewWith(v)
			case tlv.Relay:
				result.Relays = append(result.Relays, v)
			case tlv.Author:
				if len(v) < 32 {
					return prefix, nil, errorf.E("author is less than 32 bytes (%d)", len(v))
				}
				result.Author = make([]byte, schnorr.PubKeyBytesLen*2)
				hex.EncBytes(result.Author, v)
			case tlv.Kind:
				result.Kind = kind.New(binary.BigEndian.Uint32(v))
			default:
				// ignore
			}
		}
	case bytes.Equal(prefix, NentityHRP):
		var result pointers.Entity
		for {
			t, v := tlv.ReadEntry(buf)
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
			case tlv.Default:
				result.Identifier = v
			case tlv.Relay:
				result.Relays = append(result.Relays, v)
			case tlv.Author:
				if len(v) < 32 {
					return prefix, nil, errorf.E("author is less than 32 bytes (%d)", len(v))
				}
				result.PublicKey = make([]byte, schnorr.PubKeyBytesLen*2)
				hex.EncBytes(result.PublicKey, v)
			case tlv.Kind:
				result.Kind = kind.New(binary.BigEndian.Uint32(v))
			default:
				log.D.Ln("got a bogus TLV type code", t)
				// ignore
			}
		}
	}
	return prefix, data, errorf.E("unknown tag %s", prefix)
}

// EncodeNote encodes a standard nostr NIP-19 note entity (mostly meaning a
// nostr kind 1 short text note)
func EncodeNote(eventIDHex []byte) (s []byte, err error) {
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

// EncodeProfile encodes a pubkey and a set of relays into a bech32 encoded
// entity.
func EncodeProfile(publicKeyHex []byte, relays [][]byte) (s []byte, err error) {
	buf := &bytes.Buffer{}
	pb := make([]byte, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(pb, publicKeyHex); chk.D(err) {
		err = log.E.Err("invalid pubkey '%s': %w", publicKeyHex, err)
		return
	}
	tlv.WriteEntry(buf, tlv.Default, pb)
	for _, url := range relays {
		tlv.WriteEntry(buf, tlv.Relay, []byte(url))
	}
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(buf.Bytes(), 8, 5, true); chk.D(err) {
		err = log.E.Err("failed to convert bits: %w", err)
		return
	}
	return bech32.Encode(NprofileHRP, bits5)
}

// EncodeEvent encodes an event, including relay hints and author pubkey.
func EncodeEvent(eventIDHex *eventid.T, relays [][]byte, author []byte) (s []byte, err error) {
	buf := &bytes.Buffer{}
	id := make([]byte, sha256.Size)
	if _, err = hex.DecBytes(id, eventIDHex.ByteString(nil)); chk.D(err) ||
		len(id) != 32 {

		return nil, errorf.E("invalid id %d '%s': %v", len(id), eventIDHex,
			err)
	}
	tlv.WriteEntry(buf, tlv.Default, id)
	for _, url := range relays {
		tlv.WriteEntry(buf, tlv.Relay, []byte(url))
	}
	pubkey := make([]byte, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(pubkey, author); len(pubkey) == 32 {
		tlv.WriteEntry(buf, tlv.Author, pubkey)
	}
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(buf.Bytes(), 8, 5, true); chk.D(err) {
		err = log.E.Err("failed to convert bits: %w", err)
		return
	}

	return bech32.Encode(NeventHRP, bits5)
}

// EncodeEntity encodes a pubkey, kind, event Id, and relay hints.
func EncodeEntity(pk []byte, k *kind.T, id []byte, relays [][]byte) (s []byte, err error) {
	buf := &bytes.Buffer{}
	tlv.WriteEntry(buf, tlv.Default, []byte(id))
	for _, url := range relays {
		tlv.WriteEntry(buf, tlv.Relay, []byte(url))
	}
	pb := make([]byte, schnorr.PubKeyBytesLen)
	if _, err = hex.DecBytes(pb, pk); chk.D(err) {
		return nil, errorf.E("invalid pubkey '%s': %w", pb, err)
	}
	tlv.WriteEntry(buf, tlv.Author, pb)
	kindBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(kindBytes, uint32(k.ToU16()))
	tlv.WriteEntry(buf, tlv.Kind, kindBytes)
	var bits5 []byte
	if bits5, err = bech32.ConvertBits(buf.Bytes(), 8, 5, true); chk.D(err) {
		return nil, errorf.E("failed to convert bits: %w", err)
	}
	return bech32.Encode(NentityHRP, bits5)
}
