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
	RequestType() []byte
}

type Resulter interface {
	ResultType() []byte
}

type Notifier interface {
	NotificationType() []byte
}

// Implementations
//
// By embedding the following types into the message structs and writing a constructor that loads the type name,
// code can handle these without reflection, determine type via type assertion and introspect the message type via
// the interface accessor method.

type Request struct {
	Method []byte
}

func (r Request) RequestType() []byte { return r.Method }

type Response struct {
	Type []byte
	Error
}

func (r Response) ResultType() []byte { return r.Type }

type Notification struct {
	Type []byte
}

func (n Notification) NotificationType() []byte { return n.Type }

// Msat  is milli-sat, max possible value is 1000 x 21 x 100 000 000 (well, under 19 places of 64 bits in base 10)
type Msat uint64

func (m Msat) Bytes(dst []byte) (b []byte) { return ints.New(uint64(m)).Marshal(dst) }

// Methods

type Invoice struct {
	Id      []byte // nil for request, required for responses (omitted if nil)
	Invoice []byte
	Amount  Msat // optional, omitted if zero
}

type InvoiceResponse struct {
	Type            []byte // incoming or outgoing
	Invoice         []byte // optional
	Description     []byte // optional
	DescriptionHash []byte // optional
	Preimage        []byte // optional if unpaid
	PaymentHash     []byte
	Amount          Msat
	FeesPaid        Msat
	CreatedAt       int64
	ExpiresAt       int64 // optional if not applicable
	Metadata        []any // optional, probably like tags but retardation can be retarded so allow also numbers and floats

}

type ListTransactions struct {
	From   int64  // optional
	Until  int64  // optional
	Limit  int    // optional
	Offset int    // optional
	Unpaid bool   // optional default false
	Type   []byte // incoming/outgoing/empty for "both"
}

// Notifications

var (
	PaymentSent     = []byte("payment_sent")
	PaymentReceived = []byte("payment_received")
)

type PaymentSentNotification struct {
	LookupInvoiceResponse
}

type PaymentReceivedNotification struct {
	LookupInvoiceResponse
}
