package nwc

type MultiPayKeysendRequest struct {
	Request
	Keysends []PayKeysendRequest
}

func NewMultiPayKeysendRequest(keysends []PayKeysendRequest) MultiPayKeysendRequest {
	return MultiPayKeysendRequest{Request{Methods.MultiPayKeysend}, keysends}
}

type MultiPayKeysendResponse = PayKeysendResponse

func NewMultiPayKKeysendResponse(preimage []byte, feesPaid Msat) MultiPayKeysendResponse {
	return MultiPayKeysendResponse{
		Response{Type: Methods.MultiPayKeysend}, preimage, feesPaid,
	}
}
