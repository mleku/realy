package nwc

type TLV struct {
	Type  uint64
	Value by
}

type PayKeysendRequest struct {
	Request
	Amount     Msat
	Pubkey     by
	Preimage   by    // optional
	TLVRecords []TLV // optional
}

func NewPayKeysendRequest(amount Msat, pubkey, preimage by,
	tlvRecords []TLV) PayKeysendRequest {
	return PayKeysendRequest{
		Request{Methods.PayKeysend},
		amount,
		pubkey,
		preimage,
		tlvRecords,
	}
}

type PayKeysendResponse = PayInvoiceResponse

func NewPayKeysendResponse(preimage by, feesPaid Msat) PayKeysendResponse {
	return PayInvoiceResponse{
		Response{Type: Methods.PayKeysend}, preimage, feesPaid,
	}
}
