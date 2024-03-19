package contract

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	system "zcfil-server/model/system/response"
)

const (
	MinerMethodSetOperator             MethodID = "setOperator(uint64,address)"                    // Set operator
	MinerMethodMinerJoining            MethodID = "minerJoining(uint64,address)"                   // Node onboarding
	MinerMethodMinerJoiningBeneficiary MethodID = "minerJoiningBeneficiary(uint64,address)"        // Node onboarding
	MinerMethodGetMiners               MethodID = "getMiners()"                                    // Get node list
	MinerMethodGetMinerByActorId       MethodID = "getMinersByActorId(uint64)"                     // Obtain node information through node number
	MinerMethodMinerExiting            MethodID = "minerExiting(uint64)"                           // Node Resignation
	MinerMethodMinerExitingBeneficiary MethodID = "minerExitingBeneficiary(uint64)"                // Node Resignation
	MinerTotalNum                      MethodID = "minerTotal()"                                   // Number of nodes
	MinerChangeWorkerAddress           MethodID = "changeWorkerAddress(uint256,uint256,uint256[])" // Modify Worker Wallet
)

type MethodID string

func (m MethodID) Keccak256() []byte {
	return crypto.Keccak256([]byte(m))[:4]
}

type ManageContract struct {
	Contract
}

var MinerContract = ManageContract{
	Contract{
		Addr:     "0xc9c87d6C35E21637ab0B2843B626D2Dbe4EE734d",
		PushAddr: "f02815178",
	},
}

// 前端判定用
const (
	JobStatusDepart                  = iota // Not in service
	JobStatusBeOn                           // On the job
	JobStatusResigning                      // Resigned but owner remains unchanged
	JobStatusResigningNotBeneficiary        // Resigned but the beneficiary has not been changed
)

// 1. Joining, 2. Resignation, 3. Modifying operator, 4. Withdrawal, 5. Borrowing, 6. Repayment, 7. Modifying worker, 8. Modifying control, 9. Beneficiaries joining, 10. Beneficiaries leaving
const (
	OpTypeJoining = iota + 1
	OpTypeExiting
	OpTypeSetOperator
	OpTypeWithdraw
	OpTypeLoan
	OpTypeRepayment
	OpTypeChangeWorker
	OpTypeChangeControl
	OpTypeJoiningBeneficiary
	OpTypeExitingBeneficiary
)

// Database record node status
const (
	NodeStatusJoining = iota + 1 // Employment
	NodeStatusDepart             // Resignation
)

const (
	EmptyAddress = "0x0000000000000000000000000000000000000000"
)

// Operation record status
const (
	NodeActionJoining = iota + 1 // Settle in
	NodeActionChange             // Change
	NodeActionExiting            // Quit
)

func (m *ManageContract) MinerMethodGetMinerByActorId(actor string) (*system.PondActorInfo, error) {
	param := MinerMethodGetMinerByActorId.Keccak256()
	ar, _ := new(big.Int).SetString(actor, 10)
	if ar == nil {
		return nil, fmt.Errorf("Unknown node number：" + actor)
	}
	paddedActor := common.LeftPadBytes(ar.Bytes(), 32)
	param = append(param, paddedActor...)
	var pondMiner system.PondActorInfo
	return &pondMiner, m.CallContract(param, &pondMiner)
}
