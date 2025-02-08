package nwc

type MakeInvoiceRequest struct {
	Request
	Amount          Msat
	Description     []byte // optional
	DescriptionHash []byte // optional
	Expiry          int    // optional
}

func NewMakeInvoiceRequest(amount Msat, description, descriptionHash []byte,
	expiry int) MakeInvoiceRequest {
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
