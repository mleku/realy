package nwc

type LookupInvoiceRequest struct {
	Request
	PaymentHash, Invoice by
}

func NewLookupInvoiceRequest(paymentHash, invoice by) *LookupInvoiceRequest {
	return &LookupInvoiceRequest{
		Request{Methods.LookupInvoice}, paymentHash, invoice,
	}
}

type LookupInvoice struct {
	Response
	InvoiceResponse
	SettledAt int64 // optional if unpaid
}
type LookupInvoiceResponse struct {
	Response
	LookupInvoice
}

func NewLookupInvoiceResponse(resp LookupInvoice) LookupInvoiceResponse {
	return LookupInvoiceResponse{Response{Type: Methods.LookupInvoice}, resp}
}
