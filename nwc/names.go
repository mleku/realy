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
	GetInfo B
}{
	B("pay_invoice"),
	B("multi_pay_invoice"),
	B("pay_keysend"),
	B("multi_pay_keysend"),
	B("make_invoice"),
	B("lookup_invoice"),
	B("list_transactions"),
	B("get_balance"),
	B("get_info"),
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
	PaymentHash B
}{
	B("method"),
	B("params"),
	B("result_type"),
	B("error"),
	B("result"),
	B("invoice"),
	B("amount"),
	B("preimage"),
	B("fees_paid"),
	B("id"),
	B("tlv_records"),
	B("type"),
	B("value"),
	B("pubkey"),
	B("description"),
	B("description_hash"),
	B("expiry"),
	B("created_at"),
	B("expires_at"),
	B("metadata"),
	B("settled_at"),
	B("from"),
	B("until"),
	B("offset"),
	B("unpaid"),
	B("balance"),
	B("notifications"),
	B("notification_type"),
	B("notification"),
	B("payment_hash"),
}

// Notifications are the proper strings for the Notification.NotificationType
var Notifications = struct {
	PaymentReceived, PaymentSent B
}{
	B("payment_received"),
	B("payment_sent"),
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
	Other B
}{
	B("RATE_LIMITED"),
	B("NOT_IMPLEMENTED"),
	B("INSUFFICIENT_BALANCE"),
	B("QUOTA_EXCEEDED"),
	B("RESTRICTED"),
	B("UNAUTHORIZED"),
	B("INTERNAL"),
	B("OTHER"),
}
