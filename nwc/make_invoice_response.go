package nwc

type MakeInvoiceRequest struct {
	Request
	Amount          Msat
	Description     by // optional
	DescriptionHash by // optional
	Expiry          no // optional
}

func NewMakeInvoiceRequest(amount Msat, description, descriptionHash by,
	expiry no) MakeInvoiceRequest {
	return MakeInvoiceRequest{
		Request{Methods.MakeInvoice},
		amount,
		description,
		descriptionHash,
		expiry,
	}
}

type MakeInvoiceResponse struct {
	Response
	InvoiceResponse
}

func NewMakeInvoiceResponse(resp InvoiceResponse) MakeInvoiceResponse {
	return MakeInvoiceResponse{Response{Type: Methods.MakeInvoice}, resp}
}
