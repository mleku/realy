package nwc

import (
	"fmt"
)

func ExamplePayInvoiceRequest_Marshal() {
	ir := NewPayInvoiceRequest("lnbc50n1...", 0)
	var b []byte
	var err error
	if b = ir.Marshal(b); chk.E(err) {
		return
	}
	fmt.Printf("%s\n", b)
	b = b[:0]
	ir = NewPayInvoiceRequest("lnbc50n1...", 123)
	if b = ir.Marshal(b); chk.E(err) {
		return
	}
	fmt.Printf("%s\n", b)
	// Output:
	// {"method":"pay_invoice","params":{"invoice":"lnbc50n1..."}}
	// {"method":"pay_invoice","params":{"invoice":"lnbc50n1...","amount":123}}
}
