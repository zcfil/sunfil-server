package contract

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

const (
	RateGetRateContractParam MethodID = "getRateContractParam(uint256)"
	RateSetPledge            MethodID = "setPledge(uint256,uint256)"          // Set up pledge
	PoolUseRate              MethodID = "getPoolUseRate()"                    // Currency pool utilization rate
	LoanRate                 MethodID = "getLoanRate()"                       // Loan annual interest rate
	DepositRate              MethodID = "getDepositRate()"                    // Pledge annual interest rate
	RiskCoefficient          MethodID = "riskCoefficient()"                   // Risk reserve coefficient
	OmCoefficient            MethodID = "omCoefficient()"                     // Operating reserve coefficient
	MaxBorrowableAmount      MethodID = "getMaxBorrowableAmount(uint64)"      // Maximum Borrowable Limit
	MaxWithdrawalAmount      MethodID = "withdrawalLimit(uint64)"             // Maximum withdrawal limit
	RateWithdrawalLimitView  MethodID = "withdrawalLimitView(uint64,uint256)" // Withdrawal limit
)

var RateContract = Rate{
	Contract{
		Addr:     "0x2811aaa47902C87c04072921711dbe00f5445422",
		PushAddr: "f02828429",
	},
}

type Rate struct {
	Contract
}

func (m *Rate) GetMaxBorrowableAmount(actor string) *big.Int {
	var maxLoanLimit big.Int
	borrowParam := MaxBorrowableAmount.Keccak256()
	nodeId, _ := new(big.Int).SetString(actor, 10)
	paddedAmount := common.LeftPadBytes(nodeId.Bytes(), 32)
	param := append(borrowParam, paddedAmount...)
	if err := m.CallContract(param, &maxLoanLimit); err != nil {
		fmt.Println("Failed to obtain the maximum loan limit from the smart contract!", err)
	}
	return &maxLoanLimit
}
