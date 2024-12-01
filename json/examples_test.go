package json

import (
	"fmt"

	"realy.lol/hex"
)

func ExampleBool_Marshal() {
	var b by
	bt := &Bool{true}
	b = bt.Marshal(b)
	fmt.Printf("%s\n", b)
	bt2 := &Bool{}
	rem, err := bt2.Unmarshal(b)
	if err != nil || len(rem) != 0 {
		return
	}
	fmt.Printf("%v\n", bt2.V == true)
	b = b[:0]
	bf := &Bool{} // implicit initialized bool is false
	b = bf.Marshal(b)
	fmt.Printf("%s\n", b)
	fmt.Printf("%v\n", bf.V == false)
	// Output:
	// true
	// true
	// false
	// true
}

func ExampleUnsigned_Marshal() {
	var b by
	u := &Unsigned{}
	b = u.Marshal(b)
	fmt.Printf("%s\n", b)
	u2 := &Unsigned{}
	rem, err := u2.Unmarshal(b)
	if err != nil || len(rem) != 0 {
		return
	}
	fmt.Printf("%v\n", u2.V == 0)
	u.V = 69420
	b = b[:0]
	b = u.Marshal(b)
	fmt.Printf("%s\n", b)
	rem, err = u2.Unmarshal(b)
	if err != nil || len(rem) != 0 {
		return
	}
	fmt.Printf("%v\n", u2.V == 69420)
	// Output:
	// 0
	// true
	// 69420
	// true
}

func ExampleSigned_Marshal() {
	var b by
	s := &Signed{}
	b = s.Marshal(b)
	fmt.Printf("%s\n", b)
	s2 := &Signed{}
	rem, err := s2.Unmarshal(b)
	if err != nil || len(rem) != 0 {
		return
	}
	fmt.Printf("%v\n", s2.V == 0)
	s.V = 69420
	b = b[:0]
	b = s.Marshal(b)
	fmt.Printf("%s\n", b)
	rem, err = s2.Unmarshal(b)
	if err != nil || len(rem) != 0 {
		return
	}
	fmt.Printf("%v\n", s2.V == s.V)
	s.V *= -69420
	b = b[:0]
	b = s.Marshal(b)
	fmt.Printf("%s\n", b)
	rem, err = s2.Unmarshal(b)
	if err != nil || len(rem) != 0 {
		return
	}
	fmt.Printf("%v\n", s2.V == s.V)
	// Output:
	// 0
	// true
	// 69420
	// true
	// -4819136400
	// true
}

func ExampleString_Marshal() {
	var b by
	const ex = `test with
	
newlines and hidden tab and spaces at the end    `
	s := NewString(ex)
	b = s.Marshal(b)
	fmt.Printf("%s\n", b)
	s2 := &String{}
	rem, err := s2.Unmarshal(b)
	if err != nil || len(rem) != 0 {
		return
	}
	fmt.Printf("%v\n", equals(s2.V, by(ex)))
	// Output:
	// "test with\n\t\nnewlines and hidden tab and spaces at the end    "
	// true
}

func ExampleBech32_Marshal() {
	const (
		hrp    = "herp"
		hexVal = "00deadbeefcafe0123456789abcdef00deadbeefcafe0123456789abcdef"
	)
	bin, err := hex.Dec(hexVal)
	if err != nil {
		return
	}
	b32 := &Bech32{by(hrp), bin}
	b := b32.Marshal(nil)
	fmt.Printf("%s\n", b)
	b33 := &Bech32{HRP: by(hrp)}
	var rem by
	rem, err = b33.Unmarshal(b)
	if chk.E(err) || len(rem) != 0 {
		return
	}
	fmt.Printf("hrp: %s\ndata: %0x\n", b33.HRP, b33.V)
	fmt.Printf("%v\n", equals(bin, b33.V))
	// Output:
	// "herp1qr02m0h0etlqzg69v7y6hn00qr02m0h0etlqzg69v7y6hn00jujvlj"
	// hrp: herp
	// data: 00deadbeefcafe0123456789abcdef00deadbeefcafe0123456789abcdef
	// true
}

func ExampleHex_Marshal() {
	const (
		hexVal = "deadbeefcafe0123456789abcdef00deadbeefcafe0123456789abcdef"
	)
	bin, err := hex.Dec(hexVal)
	if err != nil {
		return
	}
	h := &Hex{bin}
	b := h.Marshal(nil)
	fmt.Printf("%s\n", b)
	h2 := &Hex{}
	var rem by
	rem, err = h2.Unmarshal(b)
	if chk.E(err) || len(rem) != 0 {
		fmt.Printf("%s\n%s", err.Error(), rem)
		return
	}
	fmt.Printf("data: %0x\n", h2.V)
	fmt.Printf("%v\n", equals(bin, h2.V))
	// Output:
	// "deadbeefcafe0123456789abcdef00deadbeefcafe0123456789abcdef"
	// data: deadbeefcafe0123456789abcdef00deadbeefcafe0123456789abcdef
	// true
}
func ExampleBase64_Marshal() {
	const (
		hexVal = "deadbeefcafe0123456789abcdef00deadbeefcafe0123456789abcdef00"
	)
	bin, err := hex.Dec(hexVal)
	if err != nil {
		return
	}
	b1 := &Base64{bin}
	var b by
	b = b1.Marshal(nil)
	fmt.Printf("%s\n", b)
	b2 := &Base64{}
	var rem by
	rem, err = b2.Unmarshal(b)
	if chk.E(err) || len(rem) != 0 {
		fmt.Printf("%s\n%s", err.Error(), rem)
		return
	}
	fmt.Printf("data: %0x\n", b2.V)
	fmt.Printf("%v\n", equals(bin, b2.V))
	// Output:
	// "3q2+78r+ASNFZ4mrze8A3q2+78r+ASNFZ4mrze8A"
	// data: deadbeefcafe0123456789abcdef00deadbeefcafe0123456789abcdef00
	// true
}
func ExampleKeyValue_Marshal() {
	const (
		// deliberately put whitespace here to make sure it parses. even garbage will parse, but
		// we aren't going to bother, mainly whitespace needs to be allowed.
		keyVal = `"key" :  		 
"value"`
	)
	kv := &KeyValue{Value: &String{}}
	rem, err := kv.Unmarshal(by(keyVal))
	if chk.E(err) || len(rem) != 0 {
		fmt.Printf("%s\n'%s'", err.Error(), rem)
		return
	}
	kv2 := &KeyValue{by("key"), &String{by("value")}}
	var b, b2 by
	b = kv.Marshal(b)
	b2 = kv2.Marshal(b2)
	fmt.Printf("%s\n%s\n%v\n", b, b2, equals(b, b2))
	// Output:
	// "key":"value"
	// "key":"value"
	// true
}
