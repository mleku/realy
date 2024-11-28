package nwc

import (
	"realy.lol/ints"
)

// Interfaces
//
// By using these interfaces and embedding the following implementations it becomes simple to type check the specific
// request, response or notification variable being used in a given place in the code, without using reflection.
//
// All request, responses and methods embed the implementations and their types then become easily checked.

type Requester interface {
	RequestType() B
}

type Resulter interface {
	ResultType() B
}

type Notifier interface {
	NotificationType() B
}

// Implementations
//
// By embedding the following types into the message structs and writing a constructor that loads the type name,
// code can handle these without reflection, determine type via type assertion and introspect the message type via
// the interface accessor method.

type Request struct {
	Method B
}

func (r Request) RequestType() B { return r.Method }

type Response struct {
	Type B
	Error
}

func (r Response) ResultType() B { return r.Type }

type Notification struct {
	Type B
}

func (n Notification) NotificationType() B { return n.Type }

// Msat  is milli-sat, max possible value is 1000 x 21 x 100 000 000 (well, under 19 places of 64 bits in base 10)
type Msat uint64

func (m Msat) Bytes(dst B) (b B) {
	b, _ = ints.New(uint64(m)).MarshalJSON(dst)
	return
}

// Methods

type Invoice struct {
	Id      B // nil for request, required for responses (omitted if nil)
	Invoice B
	Amount  Msat // optional, omitted if zero
}

type InvoiceResponse struct {
	Type            B // incoming or outgoing
	Invoice         B // optional
	Description     B // optional
	DescriptionHash B // optional
	Preimage        B // optional if unpaid
	PaymentHash     B
	Amount          Msat
	FeesPaid        Msat
	CreatedAt       int64
	ExpiresAt       int64 // optional if not applicable
	Metadata        []any // optional, probably like tags but retardation can be retarded so allow also numbers and floats

}

type ListTransactions struct {
	From   int64 // optional
	Until  int64 // optional
	Limit  int   // optional
	Offset int   // optional
	Unpaid bool  // optional default false
	Type   B     // incoming/outgoing/empty for "both"
}

// Notifications

var (
	PaymentSent     = B("payment_sent")
	PaymentReceived = B("payment_received")
)

type PaymentSentNotification struct {
	LookupInvoiceResponse
}

type PaymentReceivedNotification struct {
	LookupInvoiceResponse
}
