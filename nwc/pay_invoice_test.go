package nwc

import (
	"fmt"
)

func ExamplePayInvoiceRequest_MarshalJSON() {
	ir := NewPayInvoiceRequest("lnbc50n1...", 0)
	var b B
	var err E
	if b, err = ir.MarshalJSON(b); chk.E(err) {
		return
	}
	fmt.Printf("%s\n", b)
	b = b[:0]
	ir = NewPayInvoiceRequest("lnbc50n1...", 123)
	if b, err = ir.MarshalJSON(b); chk.E(err) {
		return
	}
	fmt.Printf("%s\n", b)
	// Output:
	// {"method":"pay_invoice","params":{"invoice":"lnbc50n1..."}}
	// {"method":"pay_invoice","params":{"invoice":"lnbc50n1...","amount":123}}
}
