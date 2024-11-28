package nwc

import (
	"realy.lol/text"
)

type PayInvoiceRequest struct {
	Request
	Invoice
}

func NewPayInvoiceRequest[V S | B](invoice V, amount Msat) PayInvoiceRequest {
	return PayInvoiceRequest{
		Request{Methods.PayInvoice}, Invoice{nil, B(invoice), amount},
	}
}

func (p PayInvoiceRequest) MarshalJSON(dst B) (b B, err E) {
	// open parentheses
	dst = append(dst, '{')
	// method
	dst = text.JSONKey(dst, Keys.Method)
	dst = text.Quote(dst, p.RequestType())
	dst = append(dst, ',')
	// Params
	dst = text.JSONKey(dst, Keys.Params)
	dst = append(dst, '{')
	// Invoice
	dst = text.JSONKey(dst, Keys.Invoice)
	dst = text.AppendQuote(dst, p.Invoice.Invoice, text.Noop)
	// Amount - optional (omit if zero)
	if p.Amount > 0 {
		dst = append(dst, ',')
		dst = text.JSONKey(dst, Keys.Amount)
		dst = p.Amount.Bytes(dst)
	}
	// close parentheses
	dst = append(dst, '}')
	dst = append(dst, '}')
	b = dst
	return
}

func (p PayInvoiceRequest) UnmarshalJSON(b B) (r B, err E) {

	return
}

type PayInvoiceResponse struct {
	Response
	Preimage B
	FeesPaid Msat // optional, omitted if zero
}

func NewPayInvoiceResponse(preimage B, feesPaid Msat) PayInvoiceResponse {
	return PayInvoiceResponse{
		Response{Type: Methods.PayInvoice}, preimage, feesPaid,
	}
}

func (p PayInvoiceResponse) MarshalJSON(dst B) (b B, err E) {
	// open parentheses
	dst = append(dst, '{')
	// method
	dst = text.JSONKey(dst, p.Response.Type)
	dst = text.Quote(dst, p.ResultType())
	// Params
	dst = text.JSONKey(dst, Keys.Params)
	// open parenthesis
	dst = append(dst, '{')
	// Invoice
	dst = text.JSONKey(dst, Keys.Preimage)
	dst = text.AppendQuote(dst, p.Preimage, text.Noop)
	// Amount - optional (omit if zero)
	if p.FeesPaid > 0 {
		dst = append(dst, ',')
		dst = text.JSONKey(dst, Keys.FeesPaid)
		dst = p.FeesPaid.Bytes(dst)
	}
	// close parentheses
	dst = append(dst, '}')
	dst = append(dst, '}')
	return
}

func (p PayInvoiceResponse) UnmarshalJSON(b B) (r B, err E) {
	// TODO implement me
	panic("implement me")
}
