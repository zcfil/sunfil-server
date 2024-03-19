package contract

const (
	OpNodeMethodWithdraw  MethodID = "withdraw(uint64,uint256,uint256)"  // Withdrawal
	OpNodeMethodLoan      MethodID = "loan(uint64,uint256,uint256)"      // Loan
	OpNodeMethodRepayment MethodID = "repayment(uint64,uint256,uint256)" // Repayment

)

var OpNodeContract = Contract{
	Addr:     "0xC1eb96eEc5B4c75A64f5DE36849c5f9E0B240294",
	PushAddr: "f02815179",
}
