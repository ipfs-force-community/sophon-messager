package types

type Arguments struct{}

type WalletInfo struct {
	// Parsed from the token
	WalletName string
	// Whether to compare with the wallet name in the request,
	// If the request is local(127.0.0.1), it is not required
	NeedCompare bool
}
