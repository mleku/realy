package nwc

type GetInfoRequest struct {
	Request
	// nothing to see here, move along
}

func NewGetInfoRequest() GetInfoRequest {
	return GetInfoRequest{Request{Methods.GetInfo}}
}

type GetInfo struct {
	Alias       B
	Color       B // Hex string
	Pubkey      B
	Network     B // mainnet/testnet/signet/regtest
	BlockHeight uint64
	BlockHash   B
	Methods     B // pay_invoice, get_balance, make_invoice, lookup_invoice, list_transactions, get_info (list of methods)
}

type GetInfoResponse struct {
	Response
	GetInfo
}

func NewGetInfoResponse(gi GetInfo) GetInfoResponse {
	return GetInfoResponse{Response{Type: Methods.GetInfo}, gi}
}
