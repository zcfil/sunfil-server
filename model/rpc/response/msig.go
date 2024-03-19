package response

import "github.com/filecoin-project/go-address"

type MsigInfo struct {
	Balance      string
	Spendable    string
	Threshold    uint64
	Transactions []Transact
	Signers      []AddressInfo
}

type AddressInfo struct {
	Id      string
	Address string
	Nonce   uint64
}

type Transact struct {
	TxId            int64
	Approvals       int
	ApprovalAddress []address.Address
	To              string
	Value           string
	Method          uint64
	Params          string
}
