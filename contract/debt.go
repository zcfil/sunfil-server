package contract

const (
	GetTotalSupply   MethodID = "totalSupply()"
	DebtGetBalanceOf MethodID = "balanceOf(address)"
	TotalFilBalance  MethodID = "getTotalFilBalance()"
)

var DebtContract = Contract{
	Addr: "0x67A6F45F7b180D308f3390Ba14E1F30c26C111B8",
}
