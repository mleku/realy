package nwc

type TLV struct {
	Type  uint64
	Value []byte
}

type PayKeysendRequest struct {
	Request
	Amount     Msat
	Pubkey     []byte
	Preimage   []byte // optional
	TLVRecords []TLV  // optional
}

func NewPayKeysendRequest(amount Msat, pubkey, preimage []byte,
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

func NewPayKeysendResponse(preimage []byte, feesPaid Msat) PayKeysendResponse {
	return PayInvoiceResponse{
		Response{Type: Methods.PayKeysend}, preimage, feesPaid,
	}
}
