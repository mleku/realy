package nwc

type ListTransactionsRequest struct {
	Request
	ListTransactions
}

func NewListTransactionsRequest(req ListTransactions) *ListTransactionsRequest {
	return &ListTransactionsRequest{
		Request{Methods.ListTransactions}, req,
	}
}

type ListTransactionsResponse struct {
	Response
	Transactions []LookupInvoice
}

func NewListTransactionsResponse(txs []LookupInvoice) ListTransactionsResponse {
	return ListTransactionsResponse{Response{Type: Methods.ListTransactions}, txs}
}
