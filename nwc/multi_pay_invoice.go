package nwc

type MultiPayInvoiceRequest struct {
	Request
	Invoices []Invoice
}

func NewMultiPayInvoiceRequest(invoices []Invoice) MultiPayInvoiceRequest {
	return MultiPayInvoiceRequest{
		Request:  Request{Methods.MultiPayInvoice},
		Invoices: invoices,
	}
}

type MultiPayInvoiceResponse = PayInvoiceResponse

func NewMultiPayInvoiceResponse(preimage B, feesPaid Msat) MultiPayInvoiceResponse {
	return MultiPayInvoiceResponse{Response{Type: Methods.MultiPayInvoice}, preimage, feesPaid}
}
