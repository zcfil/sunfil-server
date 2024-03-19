package internal

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-address"
	builtintypes "github.com/filecoin-project/go-state-types/builtin"
	msig11 "github.com/filecoin-project/go-state-types/builtin/v11/multisig"
	"github.com/filecoin-project/lotus/chain/actors/builtin/multisig"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	"github.com/ipfs/go-cid"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"
	"zcfil-server/contract"
	"zcfil-server/lotusrpc"
	modelSystem "zcfil-server/model/system"
	"zcfil-server/service"
	"zcfil-server/service/system"
	"zcfil-server/utils"
)

// MsgOperatorInfo Parsing Information Operations
func MsgOperatorInfo(ctx context.Context, chainCid cid.Cid, msg *types.Message, recordsMap map[string][]modelSystem.SysNodeRecords) error {
	fmt.Println("msg.Method", msg.Method)
	switch msg.Method {
	case builtintypes.MethodsEVM.InvokeContract:
		// Single signature
		if err := SingleMsg(ctx, msg, chainCid); err != nil {
			return err
		}
	case builtintypes.MethodsMultisig.Propose:
		// Multiple signature proposals
		if err := MisgProposeMsg(ctx, msg, chainCid, recordsMap); err != nil {
			return err
		}
	case builtintypes.MethodsMultisig.Approve:
		// Multiple signatures of consent
		if err := MsigApproveMsg(ctx, msg, chainCid, recordsMap); err != nil {
			return err
		}
	}
	return nil
}

// SingleMsg Single contract signing message
func SingleMsg(ctx context.Context, msg *types.Message, chainCid cid.Cid) error {
	if msg.To.String()[1:] == contract.MinerContract.Addr.ToFilecionAddress().String()[1:] || msg.To.String()[1:] == contract.MinerContract.Addr.ToActorID().String()[1:] || msg.To.String()[1:] == contract.OpNodeContract.Addr.ToFilecionAddress().String()[1:] || msg.To.String()[1:] == contract.OpNodeContract.Addr.ToActorID().String()[1:] {
		log.Println("，found contract message：", chainCid)
		record := modelSystem.SysNodeRecords{
			FromAddr:   msg.From.String(),
			ToAddr:     msg.To.String(),
			ToContract: msg.To.String(),
			Cid:        chainCid.String(),
			Amount:     msg.Value.String(),
			Applied:    true,
		}
		res, err := lotusrpc.FullApi.StateWaitMsg(ctx, chainCid, 1, -1, true)
		if err != nil {
			log.Println("SingleMsg：StateWaitMsg", err)
			return err
		}
		if res.Receipt.ExitCode.IsError() {
			log.Println("SingleMsg error message : ", chainCid, res.Receipt.ExitCode.Error())
			return err
		}

		var actor uint64
		var needUpdateDebt = false

		paramStr := utils.AnalyseParams(msg.Params)
		params, err := base64.StdEncoding.DecodeString(paramStr)
		if err != nil {
			log.Println(err, paramStr)
			params = []byte(paramStr)
		}
		param := hex.EncodeToString(params)[:8]
		log.Println("Parameter method：", string(params[:4]), "["+param+"]", "["+hex.EncodeToString(contract.OpNodeMethodLoan.Keccak256())+"]", hex.EncodeToString(contract.MinerMethodMinerExiting.Keccak256()), hex.EncodeToString(contract.MinerMethodSetOperator.Keccak256()))
		log.Println(param == hex.EncodeToString(contract.MinerMethodMinerJoining.Keccak256()))
		log.Println(param == hex.EncodeToString(contract.MinerMethodMinerExiting.Keccak256()))
		log.Println(param == hex.EncodeToString(contract.MinerMethodSetOperator.Keccak256()))
		log.Println(param == hex.EncodeToString(contract.OpNodeMethodWithdraw.Keccak256()))
		log.Println(param == hex.EncodeToString(contract.OpNodeMethodLoan.Keccak256()))
		log.Println(param == hex.EncodeToString(contract.OpNodeMethodRepayment.Keccak256()))
		switch hex.EncodeToString(params)[:8] {
		// Employment
		case hex.EncodeToString(contract.MinerMethodMinerJoining.Keccak256()), hex.EncodeToString(contract.MinerMethodMinerJoiningBeneficiary.Keccak256()):
			record.NodeAction = contract.NodeActionJoining
			var addr string
			if err = contract.MinerContract.DecodeInput(hex.EncodeToString(params), &actor, &addr); err != nil {
				log.Println("MinerMethodMinerJoining：DecodeInput", err)
				return err
			}
			filAddr, err := address.NewIDAddress(actor)
			if err != nil {
				log.Println("Failed to convert FIL address：", err, actor)
				return err
			}
			toEthAddr, err := ethtypes.EthAddressFromFilecoinAddress(filAddr)
			if err != nil {
				log.Println("Failed to convert ETH address：", err, actor)
				return err
			}
			node := modelSystem.SysNodeInfo{
				NodeName:    actor,
				Status:      contract.NodeStatusJoining,
				NodeAddress: toEthAddr.String(),
				Applied:     true,
			}

			pond, err := contract.MinerContract.MinerMethodGetMinerByActorId(strconv.FormatInt(int64(actor), 10))
			if err != nil {
				return err
			}
			if pond.OwnerId > 0 {
				node.Owner = fmt.Sprintf("f0%d", pond.OwnerId)
			} else {
				info, err := lotusrpc.FullApi.StateMinerInfo(ctx, filAddr, types.EmptyTSK)
				if err != nil {
					log.Println("Failed to obtain miner information：", err, actor)
					return err
				}
				node.Owner = info.Owner.String()
			}
			node.Operator = pond.Operator

			if err = service.ServiceGroupApp.SystemServiceGroup.CreateSysNodeInfo(node); err != nil {
				return err
			}
			if hex.EncodeToString(params)[:8] == hex.EncodeToString(contract.MinerMethodMinerJoining.Keccak256()) {
				// Owner onboarding
				record.OpType = contract.OpTypeJoining
			} else {
				// Employment of beneficiaries
				record.OpType = contract.OpTypeJoiningBeneficiary
			}
			record.OpContent = addr
		// Resignation
		case hex.EncodeToString(contract.MinerMethodMinerExiting.Keccak256()), hex.EncodeToString(contract.MinerMethodMinerExitingBeneficiary.Keccak256()):
			record.NodeAction = contract.NodeActionExiting
			if err = contract.MinerContract.DecodeInput(hex.EncodeToString(params), &actor); err != nil {
				log.Println("MinerMethodMinerExiting：DecodeInput", err)
				return err
			}

			// Modify node status
			if err = service.ServiceGroupApp.SystemServiceGroup.UpdateSysNodeStatus(actor, contract.NodeStatusDepart); err != nil {
				return err
			}

			// Modify node operator
			if err = service.ServiceGroupApp.SystemServiceGroup.UpdateSysNodeOperator(actor, contract.EmptyAddress); err != nil {
				return err
			}

			if hex.EncodeToString(params)[:8] == hex.EncodeToString(contract.MinerMethodMinerExiting.Keccak256()) {
				// Owner Resignation
				record.OpType = contract.OpTypeExiting
			} else {
				// Resignation of beneficiaries
				record.OpType = contract.OpTypeExitingBeneficiary
			}
		// Set operator
		case hex.EncodeToString(contract.MinerMethodSetOperator.Keccak256()):
			record.NodeAction = contract.NodeActionChange
			var op string
			if err = contract.MinerContract.DecodeInput(hex.EncodeToString(params), &actor, &op); err != nil {
				log.Println("MinerMethodSetOperator：DecodeInput", err)
				return err
			}
			log.Println("Get parameters：", actor, op)
			record.OpType = contract.OpTypeSetOperator
			record.OpContent = op

			// Obtain the operator and modify it
			pond, err := contract.MinerContract.MinerMethodGetMinerByActorId(strconv.FormatInt(int64(actor), 10))
			if err != nil {
				return err
			}
			if err = service.ServiceGroupApp.SystemServiceGroup.UpdateSysNodeOperator(actor, pond.Operator); err != nil {
				return err
			}
		// Withdrawal
		case hex.EncodeToString(contract.OpNodeMethodWithdraw.Keccak256()):
			var amount big.Int
			if err = contract.OpNodeContract.DecodeInput(hex.EncodeToString(params), &actor, &amount); err != nil {
				log.Println("MinerMethodWithdraw：DecodeInput", err)
				return err
			}

			info, err := contract.MinerContract.MinerMethodGetMinerByActorId(strconv.FormatUint(actor, 10))
			if err != nil {
				log.Println("MinerMethodWithdraw：MinerMethodGetMinerByActorId：", err)
			}

			addr, err := address.NewIDAddress(info.OwnerId)
			if err != nil {
				log.Println("MinerMethodWithdraw：NewIDAddress：", info.OwnerId, err)
			} else {
				if signMul, err := lotusrpc.FullApi.StateAccountKey(ctx, addr, types.EmptyTSK); err != nil {
					log.Println("MinerMethodWithdraw：StateAccountKey：", addr.String(), err)
				} else {
					addr = signMul
				}
			}

			record.OpType = contract.OpTypeWithdraw
			record.OpContent = addr.String() + "," + amount.String()
			needUpdateDebt = true
		// Loan
		case hex.EncodeToString(contract.OpNodeMethodLoan.Keccak256()):
			var amount big.Int
			if err = contract.OpNodeContract.DecodeInput(hex.EncodeToString(params), &actor, &amount); err != nil {
				log.Println("MinerMethodLoan：DecodeInput", err)
				return err
			}
			record.OpType = contract.OpTypeLoan
			needUpdateDebt = true
			record.OpContent = amount.String()
		// Repayment
		case hex.EncodeToString(contract.OpNodeMethodRepayment.Keccak256()):
			var amount big.Int
			if err = contract.OpNodeContract.DecodeInput(hex.EncodeToString(params), &actor, &amount); err != nil {
				log.Println("MinerMethodRepayment：DecodeInput", err)
				return err
			}
			record.OpType = contract.OpTypeRepayment
			record.OpContent = amount.String()
			needUpdateDebt = true
		// Modify worker
		case hex.EncodeToString(contract.MinerChangeWorkerAddress.Keccak256()):
			record.NodeAction = contract.NodeActionChange
			var worker string
			var control []string
			if err = contract.MinerContract.DecodeInput(hex.EncodeToString(params), &actor, &worker, &control); err != nil {
				log.Println("MinerChangeWorkerAddress：DecodeInput", err)
				return err
			}
			filAddr, err := address.NewIDAddress(actor)
			if err != nil {
				log.Println("Failed to convert FIL address：", err, actor)
				return err
			}
			info, err := lotusrpc.FullApi.StateMinerInfo(ctx, filAddr, types.EmptyTSK)
			if err != nil {
				log.Println("Failed to obtain miner information：", err, actor)
				return err
			}

			if worker == info.Worker.String()[2:] {
				record.OpType = contract.OpTypeChangeControl
				con, err := json.Marshal(control)
				if err != nil {
					return err
				}
				record.OpContent = string(con)
			} else {
				record.OpType = contract.OpTypeChangeWorker
				record.OpContent = worker
			}

			needUpdateDebt = true
		}
		record.ActorId = actor
		if record.ActorId != 0 {
			if err = service.ServiceGroupApp.SystemServiceGroup.CreateSysRecords(record); err != nil {
				log.Println("error：CreateSysRecords：", err.Error(), record)
				return err
			}

			// Update node debt information
			var nodeInfo modelSystem.SysNodeInfo
			nodeInfoService := system.NodeInfoService{}
			nodeInfo, err = nodeInfoService.GetSysNodeInfo(record.ActorId)
			if err != nil {
				log.Println("error：GetSysNodeInfo：", err.Error(), record)
			} else {
				if needUpdateDebt {
					UpdateNodeDebt(nodeInfo)
				}
			}
		}
	}
	return nil
}

func MisgProposeMsg(ctx context.Context, msg *types.Message, chainCid cid.Cid, recordsMap map[string][]modelSystem.SysNodeRecords) error {
	var mul multisig.ProposeParams
	if err := mul.UnmarshalCBOR(bytes.NewBuffer(msg.Params)); err != nil {
		return nil
	}
	log.Println(contract.MinerContract.Addr.ToActorID().String(), contract.MinerContract.Addr.ToFilecionAddress().String(), mul.To.String(), "Found multiple signed messages：", chainCid, mul)
	if mul.Method == builtintypes.MethodsEVM.InvokeContract && (mul.To.String()[1:] == contract.MinerContract.Addr.ToFilecionAddress().String()[1:] || mul.To.String()[1:] == contract.MinerContract.Addr.ToActorID().String()[1:]) {
		record := modelSystem.SysNodeRecords{
			FromAddr:       msg.From.String(),
			ToAddr:         msg.To.String(),
			Cid:            chainCid.String(),
			Amount:         msg.Value.String(),
			ToContract:     mul.To.String(),
			AmountContract: mul.Value.String(),
			Applied:        false,
			IsMultisig:     true,
		}

		paramStr := utils.AnalyseParams(mul.Params)
		params, err := base64.StdEncoding.DecodeString(paramStr)
		if err != nil {
			log.Println(err, paramStr)
			params = []byte(paramStr)
		}
		var actor uint64
		switch hex.EncodeToString(params)[:8] {
		// Employment
		case hex.EncodeToString(contract.MinerMethodMinerJoining.Keccak256()), hex.EncodeToString(contract.MinerMethodMinerJoiningBeneficiary.Keccak256()):
			record.NodeAction = contract.NodeActionJoining
			var addr string
			if err = contract.MinerContract.DecodeInput(hex.EncodeToString(params), &actor, &addr); err != nil {
				log.Println("MinerMethodMinerJoining：DecodeInput", err)
				return err
			}
			// Analyzing Internal Transactions in Contracts
			if err = decodeProposeMultisig(ctx, chainCid, &record); err != nil {
				return err
			}
			filAddr, err := address.NewIDAddress(actor)
			if err != nil {
				log.Println("Failed to convert FIL address：", err, actor)
				return err
			}
			toEthAddr, err := ethtypes.EthAddressFromFilecoinAddress(filAddr)
			if err != nil {
				log.Println("Failed to convert ETH address：", err, actor)
				return err
			}

			node := modelSystem.SysNodeInfo{
				NodeName:    actor,
				Status:      contract.NodeStatusJoining,
				NodeAddress: toEthAddr.String(),
				Applied:     record.Applied,
			}
			// Obtain the operator and assign a value
			pond, err := contract.MinerContract.MinerMethodGetMinerByActorId(strconv.FormatInt(int64(actor), 10))
			if err != nil {
				return err
			}
			if pond.OwnerId > 0 {
				node.Owner = fmt.Sprintf("f0%d", pond.OwnerId)
			} else {
				info, err := lotusrpc.FullApi.StateMinerInfo(ctx, filAddr, types.EmptyTSK)
				if err != nil {
					log.Println("Failed to obtain miner information：", err, actor)
					return err
				}
				node.Owner = info.Owner.String()
			}
			node.Operator = pond.Operator
			if err = service.ServiceGroupApp.SystemServiceGroup.CreateSysNodeInfo(node); err != nil {
				return err
			}

			if hex.EncodeToString(params)[:8] == hex.EncodeToString(contract.MinerMethodMinerJoining.Keccak256()) {
				record.OpType = contract.OpTypeJoining
			} else {
				record.OpType = contract.OpTypeJoiningBeneficiary
			}
			record.OpContent = addr
		// Resignation
		case hex.EncodeToString(contract.MinerMethodMinerExiting.Keccak256()), hex.EncodeToString(contract.MinerMethodMinerExitingBeneficiary.Keccak256()):
			record.NodeAction = contract.NodeActionExiting
			if err = contract.MinerContract.DecodeInput(hex.EncodeToString(params), &actor); err != nil {
				log.Println("MinerMethodMinerExiting：DecodeInput", err)
				return err
			}
			// Analyzing Internal Transactions in Contracts
			if err = decodeProposeMultisig(ctx, chainCid, &record); err != nil {
				return err
			}
			if hex.EncodeToString(params)[:8] == hex.EncodeToString(contract.MinerMethodMinerExiting.Keccak256()) {
				record.OpType = contract.OpTypeExiting
			} else {
				record.OpType = contract.OpTypeExitingBeneficiary
			}
			if record.Applied {
				if err = service.ServiceGroupApp.SystemServiceGroup.UpdateSysNodeStatus(actor, contract.NodeStatusDepart); err != nil {
					return err
				}
				if err = service.ServiceGroupApp.SystemServiceGroup.UpdateSysNodeOperator(actor, contract.EmptyAddress); err != nil {
					return err
				}
			}
		// Set operator
		case hex.EncodeToString(contract.MinerMethodSetOperator.Keccak256()):
			record.NodeAction = contract.NodeActionChange
			var op string
			if err = contract.MinerContract.DecodeInput(hex.EncodeToString(params), &actor, &op); err != nil {
				log.Println("MinerMethodSetOperator：DecodeInput", err)
				return err
			}
			// Analyzing Internal Transactions in Contracts
			if err = decodeProposeMultisig(ctx, chainCid, &record); err != nil {
				return err
			}

			record.OpType = contract.OpTypeSetOperator
			record.OpContent = op
			if record.Applied {
				if err = service.ServiceGroupApp.SystemServiceGroup.UpdateSysNodeOperator(actor, op); err != nil {
					return err
				}
			}
		// Modify worker
		case hex.EncodeToString(contract.MinerChangeWorkerAddress.Keccak256()):
			record.NodeAction = contract.NodeActionChange
			var worker string
			var control []string
			if err = contract.MinerContract.DecodeInput(hex.EncodeToString(params), &actor, &worker, &control); err != nil {
				log.Println("MinerChangeWorkerAddress：DecodeInput", err)
				return err
			}
			filAddr, err := address.NewIDAddress(actor)
			if err != nil {
				log.Println("Failed to convert FIL address：", err, actor)
				return err
			}
			// Analyzing Internal Transactions in Contracts
			if err = decodeProposeMultisig(ctx, chainCid, &record); err != nil {
				return err
			}
			info, err := lotusrpc.FullApi.StateMinerInfo(ctx, filAddr, types.EmptyTSK)
			if err != nil {
				log.Println("Failed to obtain miner information：", err, actor)
				return err
			}

			if worker == info.Worker.String()[2:] {
				record.OpType = contract.OpTypeChangeControl
				con, err := json.Marshal(control)
				if err != nil {
					return err
				}
				record.OpContent = string(con)
			} else {
				record.OpType = contract.OpTypeChangeWorker
				record.OpContent = worker
			}
		}
		if actor != 0 {
			record.ActorId = actor
			if err = service.ServiceGroupApp.SystemServiceGroup.CreateSysRecords(record); err != nil {
				log.Println("error：CreateSysRecords：", err.Error(), record)
				return err
			}
			if !record.Applied {
				recordsMap[msg.To.String()[1:]] = append(recordsMap[msg.To.String()[1:]], record)
			}
		}
	}

	return nil
}

func MsigApproveMsg(ctx context.Context, msg *types.Message, chainCid cid.Cid, recordsMap map[string][]modelSystem.SysNodeRecords) error {
	to, err := lotusrpc.FullApi.StateLookupID(ctx, msg.To, types.EmptyTSK)
	if err != nil {
		return err
	}
	toAddr := to.String()[1:]
	records := recordsMap[toAddr]
	if records == nil {
		records = recordsMap[msg.To.String()[1:]]
	}
	if records != nil {
		var txnParam msig11.TxnIDParams
		if err := txnParam.UnmarshalCBOR(bytes.NewBuffer(msg.Params)); err != nil {
			log.Println("txnParam.UnmarshalCBOR error:", err)
			return err
		}
		log.Println(msg.To.String(), "Found multiple app messages：", chainCid, txnParam)

		applied, _ := decodeAppliedMultisig(ctx, chainCid)
		if applied {
			for i, record := range records {
				if record.TxId == int64(txnParam.ID) {
					log.Println("Find the application：", msg, record.TxId)
					record.UpdatedAt = time.Now()
					record.Applied = true
					if err = service.ServiceGroupApp.SystemServiceGroup.UpdateSysRecords(record); err != nil {
						log.Println("error：UpdateSysRecords：", err.Error(), record)
						return err
					}
					switch record.OpType {
					case contract.OpTypeJoining, contract.OpTypeJoiningBeneficiary:
						// Entry node data takes effect
						if err = service.ServiceGroupApp.SystemServiceGroup.UpdateSysNodeApplied(record.ActorId, true); err != nil {
							log.Println("error：UpdateSysRecords：", err.Error(), record)
							return err
						}
					case contract.OpTypeExiting, contract.OpTypeExitingBeneficiary:
						// Change in Resignation Status
						if err = service.ServiceGroupApp.SystemServiceGroup.UpdateSysNodeStatus(record.ActorId, contract.NodeStatusDepart); err != nil {
							return err
						}
						// Modify node operator
						if err = service.ServiceGroupApp.SystemServiceGroup.UpdateSysNodeOperator(record.ActorId, contract.EmptyAddress); err != nil {
							return err
						}
					}

					recordsMap[toAddr][i] = recordsMap[toAddr][len(recordsMap[toAddr])-1]
					recordsMap[toAddr] = recordsMap[toAddr][:len(recordsMap[toAddr])-1]
					return nil
				}
			}
		}
	}
	return nil
}

func UpdateNodeDebt(nodeInfo modelSystem.SysNodeInfo) {
	var debtBalance big.Int
	param := contract.DebtGetBalanceOf.Keccak256()
	addr := common.HexToAddress(nodeInfo.NodeAddress)
	paramData := common.LeftPadBytes(addr.Bytes(), 32)
	param = append(param, paramData...)
	if err := contract.DebtContract.CallContract(param, &debtBalance); err != nil {
		fmt.Println("Failed to retrieve recorded loan pool amount from smart contract!", err)
		return
	}

	_debtBalance, _ := strconv.ParseFloat(debtBalance.String(), 64)
	_debtBalance = _debtBalance / math.Pow(10, 18)

	nodeInfo.DebtBalance = fmt.Sprintf("%.2f", _debtBalance)
	nodeAddr, err := address.NewFromString("f0" + strconv.FormatInt(int64(nodeInfo.NodeName), 10))
	if err != nil {
		fmt.Println("Node number processing failed!", err)
	} else {
		nodeInfo.Balance, err = GetMinerBalance(nodeAddr)
		if err != nil {
			fmt.Println("Failed to obtain node total amount!", err)
		}
	}

	nodeInfoService := system.NodeInfoService{}
	err = nodeInfoService.UpdateNodeInfo(&nodeInfo)
	if err != nil {
		log.Println("Node "+strconv.FormatInt(int64(nodeInfo.NodeName), 10)+" debt information and node total amount data recording failed!", err.Error())
		return
	}
}

func decodeProposeMultisig(ctx context.Context, cid cid.Cid, record *modelSystem.SysNodeRecords) error {
	res, err := lotusrpc.FullApi.StateWaitMsg(ctx, cid, 1, -1, true)
	if err != nil {
		log.Println("decodeProposeMultisig：StateWaitMsg", err)
		return err
	}
	if res.Receipt.ExitCode.IsError() {
		log.Println("propose error message : ", res.Receipt.ExitCode.Error())
		return err
	}
	var retval multisig.ProposeReturn
	if err = retval.UnmarshalCBOR(bytes.NewReader(res.Receipt.Return)); err != nil {
		return fmt.Errorf("failed to unmarshal propose return value: %w", err)
	}

	record.TxId = int64(retval.TxnID)
	record.Applied = retval.Applied
	return nil
}

func decodeAppliedMultisig(ctx context.Context, cid cid.Cid) (bool, error) {
	res, err := lotusrpc.FullApi.StateWaitMsg(ctx, cid, 1, -1, true)
	if err != nil {
		log.Println("decodeAppliedMultisig：StateWaitMsg", err)
		return false, err
	}
	if res.Receipt.ExitCode.IsError() {
		log.Println("applied error message : ", res.Receipt.ExitCode.Error())
		return false, err
	}
	var retval multisig.ApproveReturn
	if err = retval.UnmarshalCBOR(bytes.NewReader(res.Receipt.Return)); err != nil {
		return false, fmt.Errorf("failed to unmarshal propose return value: %w", err)
	}

	return retval.Applied, nil
}

// GetMinerBalance Obtain node quota
func GetMinerBalance(maddr address.Address) (float64, error) {
	var ctx = context.Background()
	mact, err := lotusrpc.FullApi.StateGetActor(ctx, maddr, types.EmptyTSK)
	if err != nil {
		return 0, err
	}

	balance := mact.Balance.String()
	nodeBalance, _ := strconv.ParseFloat(balance, 64)
	return nodeBalance, nil
}
