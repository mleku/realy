package nwc

import (
	"realy.lol/kind"
)

var Kinds = []*kind.T{
	kind.WalletInfo,
	kind.WalletRequest,
	kind.WalletResponse,
	kind.WalletNotification,
}

type Server struct {
}

type Client struct {
}
