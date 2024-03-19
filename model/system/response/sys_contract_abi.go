package response

type SysContractAbi struct {
	AbiId      string // Abi ID
	AbiName    string // Abi name
	ActorID    string // Contract MinerId
	Wallet     string // Sending a message wallet
	AbiContent string // Abi Content
	Param      []byte // Parameter
}
