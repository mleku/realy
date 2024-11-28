package nwc

type GetBalanceRequest struct {
	Request
	// nothing to see here, move along
}

func NewGetBalanceRequest() *GetBalanceRequest {
	return &GetBalanceRequest{Request{Methods.GetBalance}}
}

type GetBalanceResponse struct {
	Response
	Balance Msat
}

func NewGetBalanceResponse(balance Msat) *GetBalanceResponse {
	return &GetBalanceResponse{Response{Type: Methods.GetBalance}, balance}
}
