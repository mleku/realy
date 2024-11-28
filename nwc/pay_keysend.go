package nwc

type TLV struct {
	Type  uint64
	Value B
}

type PayKeysendRequest struct {
	Request
	Amount     Msat
	Pubkey     B
	Preimage   B     // optional
	TLVRecords []TLV // optional
}

func NewPayKeysendRequest(amount Msat, pubkey, preimage B, tlvRecords []TLV) PayKeysendRequest {
	return PayKeysendRequest{
		Request{Methods.PayKeysend},
		amount,
		pubkey,
		preimage,
		tlvRecords,
	}
}

type PayKeysendResponse = PayInvoiceResponse

func NewPayKeysendResponse(preimage B, feesPaid Msat) PayKeysendResponse {
	return PayInvoiceResponse{
		Response{Type: Methods.PayKeysend}, preimage, feesPaid,
	}
}
