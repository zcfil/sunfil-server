package define

const (
	ContractVoteId = "10000"
	ContractMiner  = "10001"
	ContractRateId = "10002"
	ContractTest   = "10004"
	ContractMethod = "3844450837"
)

var (
	NodeRecordBorrow      = "5"                                          // Loan
	NodeRecordRepayment   = "6"                                          // Repayment
	NodeRecordWithdraw    = "4"                                          // Withdrawal
	NodeRecordChange      = []string{"1", "2", "3", "7", "8", "9", "10"} // Node change
	NodeRecordLiquidation = []string{"11", "12"}                         // Node clearing
)

var (
	NodeRecordBorrowStr      = "borrow"
	NodeRecordRepaymentStr   = "repayment"
	NodeRecordWithdrawStr    = "withdraw"
	NodeRecordChangeStr      = "nodeChange"
	NodeRecordLiquidationStr = "liquidation"
)

// Cache key related
var (
	StakeAPYKey             = "stakeAPY"             // Pledged interest rate
	LoanAPYKey              = "loanAPY"              // Loan interest rate
	FinanceUseRateKey       = "financeUseRate"       // Currency pool utilization rate
	TotalBalanceKey         = "totalBalance"         // Total pool volume
	LastAvailableBalanceKey = "lastAvailableBalance" // Remaining available current assets
	StakeNumKey             = "stakeNum"             // Number of pledged individuals
	RiskCoefficientKey      = "riskCoefficient"      // Risk reserve coefficient
	BusinessCoefficientKey  = "businessCoefficient"  // Operating reserve coefficient
	TotalLoanKey            = "totalLoan"            // Total loan amount
	TotalStakeKey           = "totalStake"           // Total pledged amount

	FilCoinPriceKey = "filCoinPrice" // FIL price
)
