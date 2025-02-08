package nwc

// Methods are the text of the value of the Method field of Request.Method and
// Response.ResultType in a form that allows more convenient reference than using
// a map or package scoped variable. These appear in the API Request and Response
// types.
var Methods = struct {
	PayInvoice,
	MultiPayInvoice,
	PayKeysend,
	MultiPayKeysend,
	MakeInvoice,
	LookupInvoice,
	ListTransactions,
	GetBalance,
	GetInfo []byte
}{
	[]byte("pay_invoice"),
	[]byte("multi_pay_invoice"),
	[]byte("pay_keysend"),
	[]byte("multi_pay_keysend"),
	[]byte("make_invoice"),
	[]byte("lookup_invoice"),
	[]byte("list_transactions"),
	[]byte("get_balance"),
	[]byte("get_info"),
}

// Keys are the proper JSON bytes for the JSON object keys of the structs of the
// same-named type used lower in the following. Anonymous struct syntax is used
// to make neater addressing of these fields as symbols.
var Keys = struct {
	Method,
	Params,
	ResultType,
	Error,
	Result,
	Invoice,
	Amount,
	Preimage,
	FeesPaid,
	Id,
	TLVRecords,
	Type,
	Value,
	Pubkey,
	Description,
	DescriptionHash,
	Expiry,
	CreatedAt,
	ExpiresAt,
	Metadata,
	SettledAt,
	From,
	Until,
	Offset,
	Unpaid,
	Balance,
	Notifications,
	NotificationType,
	Notification,
	PaymentHash []byte
}{
	[]byte("method"),
	[]byte("params"),
	[]byte("result_type"),
	[]byte("error"),
	[]byte("result"),
	[]byte("invoice"),
	[]byte("amount"),
	[]byte("preimage"),
	[]byte("fees_paid"),
	[]byte("id"),
	[]byte("tlv_records"),
	[]byte("type"),
	[]byte("value"),
	[]byte("pubkey"),
	[]byte("description"),
	[]byte("description_hash"),
	[]byte("expiry"),
	[]byte("created_at"),
	[]byte("expires_at"),
	[]byte("metadata"),
	[]byte("settled_at"),
	[]byte("from"),
	[]byte("until"),
	[]byte("offset"),
	[]byte("unpaid"),
	[]byte("balance"),
	[]byte("notifications"),
	[]byte("notification_type"),
	[]byte("notification"),
	[]byte("payment_hash"),
}

// Notifications are the proper strings for the Notification.NotificationType
var Notifications = struct {
	PaymentReceived, PaymentSent []byte
}{
	[]byte("payment_received"),
	[]byte("payment_sent"),
}

var Errors = struct {
	// RateLimited - The client is sending commands too fast.It should retry in a few seconds.
	RateLimited,
	// NotImplemented - The command is not known or is intentionally not implemented.
	NotImplemented,
	// InsufficientBalance - The wallet does not have enough funds to cover a fee reserve or the payment amount.
	InsufficientBalance,
	// QuotaExceeded - The wallet has exceeded its spending quota.
	QuotaExceeded,
	// Restricted - This public key is not allowed to do this operation.
	Restricted,
	// Unauthorized - This public key has no wallet connected.
	Unauthorized,
	// Internal - An internal error.
	Internal,
	// Other - Other error.
	Other []byte
}{
	[]byte("RATE_LIMITED"),
	[]byte("NOT_IMPLEMENTED"),
	[]byte("INSUFFICIENT_BALANCE"),
	[]byte("QUOTA_EXCEEDED"),
	[]byte("RESTRICTED"),
	[]byte("UNAUTHORIZED"),
	[]byte("INTERNAL"),
	[]byte("OTHER"),
}
