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
	GetInfo by
}{
	by("pay_invoice"),
	by("multi_pay_invoice"),
	by("pay_keysend"),
	by("multi_pay_keysend"),
	by("make_invoice"),
	by("lookup_invoice"),
	by("list_transactions"),
	by("get_balance"),
	by("get_info"),
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
	PaymentHash by
}{
	by("method"),
	by("params"),
	by("result_type"),
	by("error"),
	by("result"),
	by("invoice"),
	by("amount"),
	by("preimage"),
	by("fees_paid"),
	by("id"),
	by("tlv_records"),
	by("type"),
	by("value"),
	by("pubkey"),
	by("description"),
	by("description_hash"),
	by("expiry"),
	by("created_at"),
	by("expires_at"),
	by("metadata"),
	by("settled_at"),
	by("from"),
	by("until"),
	by("offset"),
	by("unpaid"),
	by("balance"),
	by("notifications"),
	by("notification_type"),
	by("notification"),
	by("payment_hash"),
}

// Notifications are the proper strings for the Notification.NotificationType
var Notifications = struct {
	PaymentReceived, PaymentSent by
}{
	by("payment_received"),
	by("payment_sent"),
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
	Other by
}{
	by("RATE_LIMITED"),
	by("NOT_IMPLEMENTED"),
	by("INSUFFICIENT_BALANCE"),
	by("QUOTA_EXCEEDED"),
	by("RESTRICTED"),
	by("UNAUTHORIZED"),
	by("INTERNAL"),
	by("OTHER"),
}
