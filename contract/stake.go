package contract

const (
	GetContractAmount  MethodID = "getContractAmount()" // Obtain the balance of node smart contracts
	GetTotalFilBalance MethodID = "getTotalFilBalance()"
	GetFinanceUseRate  MethodID = "getFinanceUseRate()"  // Fund utilization rate
	StakeAddressNum    MethodID = "getStakeAddressNum()" // Number of pledged individuals
	StakeTotalSupply   MethodID = "totalSupply()"        // Total amount pledged
	AddressBalance     MethodID = "balanceOf(address)"   // The amount of currency corresponding to the address
)

var StakeContract = Contract{
	Addr: "0x1a1994ddF4C0c26Ee57823b3Ca2c232F8d268935",
}
