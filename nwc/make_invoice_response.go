package nwc

type MakeInvoiceRequest struct {
	Request
	Amount          Msat
	Description     B // optional
	DescriptionHash B // optional
	Expiry          N // optional
}

func NewMakeInvoiceRequest(amount Msat, description, descriptionHash B,
	expiry N) MakeInvoiceRequest {
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
