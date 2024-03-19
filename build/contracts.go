package build

import (
	_ "embed"
	"zcfil-server/contract"
)

//go:embed contract-json/miner.json
var MinerContracts string

//go:embed contract-json/op.json
var OpNodeContracts string

//go:embed contract-json/vote.json
var VoteContracts string

//go:embed contract-json/rate.json
var RateContracts string

//go:embed contract-json/stake.json
var StakeContracts string

//go:embed contract-json/debt.json
var DebtContracts string

func InitContracts() {
	contract.MinerContract.SetABI(MinerContracts)
	contract.VoteContract.SetABI(VoteContracts)
	contract.RateContract.SetABI(RateContracts)
	contract.StakeContract.SetABI(StakeContracts)
	contract.DebtContract.SetABI(DebtContracts)
	contract.OpNodeContract.SetABI(OpNodeContracts)
}
