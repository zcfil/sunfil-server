package system

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/actors/builtin/multisig"
	"github.com/filecoin-project/lotus/chain/consensus"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	cbor "github.com/ipfs/go-ipld-cbor"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"zcfil-server/contract"
	"zcfil-server/lotusrpc"
	"zcfil-server/model/rpc/response"
	system "zcfil-server/model/system/response"
)

var NodeQueryServiceApi = new(NodeQueryService)

type NodeQueryService struct{}

// MsigInspect Obtain signature information inspect
func (l *NodeQueryService) MsigInspect(ctx context.Context, actor address.Address) (*response.MsigInfo, error) {
	store := adt.WrapStore(ctx, cbor.NewCborStore(blockstore.NewAPIBlockstore(lotusrpc.FullApi)))

	head, err := lotusrpc.FullApi.ChainHead(ctx)
	if err != nil {
		return nil, err
	}

	act, err := lotusrpc.FullApi.StateGetActor(ctx, actor, head.Key())
	if err != nil {
		return nil, err
	}

	ownId, err := lotusrpc.FullApi.StateLookupID(ctx, actor, types.EmptyTSK)
	if err != nil {
		return nil, err
	}
	log.Println(act.Code.String())
	mstate, err := multisig.Load(store, act)
	if err != nil {
		return nil, err
	}
	locked, err := mstate.LockedBalance(head.Height())
	if err != nil {
		return nil, err
	}
	var res response.MsigInfo
	// balance
	res.Balance = act.Balance.String()
	res.Spendable = types.BigSub(act.Balance, locked).String()

	signers, err := mstate.Signers()
	if err != nil {
		return nil, err
	}
	threshold, err := mstate.Threshold()
	if err != nil {
		return nil, err
	}
	// By threshold
	res.Threshold = threshold

	// Can concurrently obtain
	for _, s := range signers {
		var addr response.AddressInfo
		addr.Id = s.String()
		signerActor, err := lotusrpc.FullApi.StateAccountKey(ctx, s, types.EmptyTSK)
		if err != nil {
			return nil, err
		}
		addr.Address = signerActor.String()
		nonce, err := lotusrpc.FullApi.MpoolGetNonce(ctx, signerActor)
		if err != nil {
			return nil, err
		}
		addr.Nonce = nonce
		res.Signers = append(res.Signers, addr)
	}

	pending := make(map[int64]multisig.Transaction)
	if err := mstate.ForEachPendingTxn(func(id int64, txn multisig.Transaction) error {
		pending[id] = txn
		return nil
	}); err != nil {
		return nil, err
	}

	if len(pending) > 0 {
		var txids []int64
		for txid := range pending {
			txids = append(txids, txid)
		}
		sort.Slice(txids, func(i, j int) bool {
			return txids[i] < txids[j]
		})
		methods := consensus.NewActorRegistry()
		for _, txid := range txids {
			tx := pending[txid]
			target := tx.To.String()
			if tx.To == ownId {
				target += " (self)"
			}
			paramStr := fmt.Sprintf("%x", tx.Params)
			if err != nil {
				if tx.Method == 0 {
					fmt.Printf("%d\t%s\t%d\t%s\t%s\t%s(%d)\t%s\n", txid, "pending", len(tx.Approved), target, types.FIL(tx.Value), "Send", tx.Method, paramStr)
				} else {
					fmt.Printf("%d\t%s\t%d\t%s\t%s\t%s(%d)\t%s\n", txid, "pending", len(tx.Approved), target, types.FIL(tx.Value), "new account, unknown method", tx.Method, paramStr)
				}
			} else {
				for _, v := range methods.Methods {
					method := v[tx.Method]
					if tx.Method != 0 && method.Params != nil {
						ptyp := reflect.New(method.Params.Elem()).Interface().(cbg.CBORUnmarshaler)
						if err := ptyp.UnmarshalCBOR(bytes.NewReader(tx.Params)); err != nil {
							//log.Println("failed to decode parameters of transaction ", txid, err)
							continue
						}

						b, err := json.Marshal(ptyp)
						if err != nil {
							continue
							//return xerrors.Errorf("could not json marshal parameter type: %w", err)
						}
						if string(b) == "{}" {
							continue
						}
						paramStr = strings.ReplaceAll(string(b), `"`, "")
						break
					}
				}
				tran := response.Transact{
					txid,
					len(tx.Approved),
					tx.Approved,
					target,
					tx.Value.String(),
					uint64(tx.Method),
					paramStr,
				}
				res.Transactions = append(res.Transactions, tran)
			}
		}

	}
	return &res, nil
}

// HandleMinerInfo Obtain miner information
func (l *NodeQueryService) HandleMinerInfo(maddr address.Address) (res response.MinerInfo, err error) {
	var ctx = context.Background()
	mact, err := lotusrpc.FullApi.StateGetActor(ctx, maddr, types.EmptyTSK)
	if err != nil {
		return response.MinerInfo{}, err
	}

	tbs := blockstore.NewTieredBstore(blockstore.NewAPIBlockstore(lotusrpc.FullApi), blockstore.NewMemory())

	mas, err := miner.Load(adt.WrapStore(ctx, cbor.NewCborStore(tbs)), mact)
	if err != nil {
		return response.MinerInfo{}, err
	}

	// Sector size
	mi, err := lotusrpc.FullApi.StateMinerInfo(ctx, maddr, types.EmptyTSK)
	if err != nil {
		return response.MinerInfo{}, err
	}

	res.SectorSize = types.SizeStr(types.NewInt(uint64(mi.SectorSize)))
	res.OwnerId = mi.Owner.String()
	if owner, err := lotusrpc.FullApi.StateAccountKey(ctx, mi.Owner, types.EmptyTSK); err == nil {
		res.Owner = owner.String()
	} else {
		res.Owner = mi.Owner.String()
	}
	pond, err := contract.MinerContract.MinerMethodGetMinerByActorId(maddr.String()[2:])
	if err == nil && pond.OwnerId != 0 {
		res.MortgageType = pond.MortgageType
		addr, err := address.NewIDAddress(pond.OwnerId)
		if err != nil {
			return response.MinerInfo{}, err
		}
		if oldOwner, err := lotusrpc.FullApi.StateAccountKey(ctx, addr, types.EmptyTSK); err == nil {
			res.OldOwner = oldOwner.String()
		} else {
			res.OldOwner = addr.String()
		}
	}

	res.Beneficiary = mi.Beneficiary.String()
	if bene, err := lotusrpc.FullApi.StateAccountKey(ctx, mi.Beneficiary, types.EmptyTSK); err == nil {
		res.Beneficiary = bene.String()
	} else {
		res.BeneficiaryId = mi.Beneficiary.String()
	}

	pow, err := lotusrpc.FullApi.StateMinerPower(ctx, maddr, types.EmptyTSK)
	if err != nil {
		return response.MinerInfo{}, err
	}
	res.Power = types.SizeStr(pow.MinerPower.QualityAdjPower)
	secCounts, err := lotusrpc.FullApi.StateMinerSectorCount(ctx, maddr, types.EmptyTSK)
	if err != nil {
		return response.MinerInfo{}, err
	}
	res.Active = secCounts.Active
	res.Faulty = secCounts.Faulty
	res.Live = secCounts.Live

	// NOTE: there's no need to unlock anything here. Funds only
	// vest on deadline boundaries, and they're unlocked by cron.
	lockedFunds, err := mas.LockedFunds()
	if err != nil {
		return response.MinerInfo{}, xerrors.Errorf("Getting locked funds: %w", err)
	}
	// Locking, pledging, and node balance
	res.Balance = mact.Balance
	res.Pledge = types.FIL(lockedFunds.InitialPledgeRequirement).Short()
	res.Vesting = types.FIL(lockedFunds.VestingFunds).Short()
	// Available balance
	res.Available, err = mas.AvailableBalance(mact.Balance)
	if err != nil {
		return response.MinerInfo{}, xerrors.Errorf("Getting available balance: %w", err)
	}
	// Maximum Borrowable Limit
	res.BorrowableAmount = types.BigInt{contract.RateContract.GetMaxBorrowableAmount(maddr.String()[2:])}
	return res, nil
}

// HandleWorkerInfo Obtain worker information
func (l *NodeQueryService) HandleWorkerInfo(maddr address.Address) (res response.WorkerInfo, err error) {
	var ctx = context.Background()
	// Sector size
	mi, err := lotusrpc.FullApi.StateMinerInfo(ctx, maddr, types.EmptyTSK)
	if err != nil {
		return res, err
	}
	addr := strings.Replace(maddr.String(), "f0", "", 1)
	addr = strings.Replace(addr, "t0", "", 1)
	res.Worker = mi.Worker.String()
	res.ControlAddresses = mi.ControlAddresses
	res.WorkerChangeEpoch = int64(mi.WorkerChangeEpoch)
	res.NewWorker = mi.NewWorker.String()
	pond, err := contract.MinerContract.MinerMethodGetMinerByActorId(addr)
	if err != nil {
		log.Println(err)
		return response.WorkerInfo{}, fmt.Errorf("Node not hired！")
	}
	owner, err := address.NewIDAddress(pond.OwnerId)
	if err != nil {
		return response.WorkerInfo{}, err
	}
	res.Owner = owner.String()
	ownerAddr, err := lotusrpc.FullApi.StateAccountKey(ctx, owner, types.EmptyTSK)
	if err == nil {
		res.Owner = ownerAddr.String()
	}
	return res, nil
}

// GetOperator Get operator
func (l *NodeQueryService) GetOperator(ctx context.Context, addr address.Address) (res system.ActorInfo, err error) {
	actor := addr.String()
	res.Actor = actor
	actor = strings.Replace(actor, "t0", "", 1)
	actor = strings.Replace(actor, "f0", "", 1)

	// Get operator
	param, err := contract.MinerContract.MinerMethodGetMinerByActorId(actor)

	res.Operator = param.Operator

	// Get the owner wallet
	minerInfo, err := contract.MinerContract.MinerMethodGetMinerByActorId(actor)
	if err != nil {
		return res, err
	}

	ownerId, err := address.NewIDAddress(minerInfo.OwnerId)
	if err != nil {
		return res, err
	}
	res.Owner = ownerId.String()
	ownerAddr, err := lotusrpc.FullApi.StateAccountKey(ctx, ownerId, types.EmptyTSK)
	if err == nil {
		res.Owner = ownerAddr.String()
	}

	//owner nonce info
	res.Nonce, _ = lotusrpc.FullApi.MpoolGetNonce(ctx, ownerAddr)
	return res, nil
}

// DepartInfo Obtain resignation information
func (l *NodeQueryService) DepartInfo(ctx context.Context, addr address.Address) (res system.PondActorStatus, err error) {
	actor := addr.String()
	res.Actor = actor
	actor = strings.Replace(actor, "t0", "", 1)
	actor = strings.Replace(actor, "f0", "", 1)

	mi, err := lotusrpc.FullApi.StateMinerInfo(ctx, addr, types.EmptyTSK)
	if err != nil {
		return res, err
	}
	owner, err := lotusrpc.FullApi.StateAccountKey(ctx, mi.Owner, types.EmptyTSK)
	if err != nil {
		res.Owner = mi.Owner.String()
	} else {
		res.Owner = owner.String()
	}

	res.Beneficiary = mi.Beneficiary.String()

	contractAddr := contract.SunContract.Addr.ToActorID()
	fmt.Println(mi.Owner, contractAddr, mi.Beneficiary)

	pondMiner, _ := contract.MinerContract.MinerMethodGetMinerByActorId(actor)
	log.Println("合约地址：", contractAddr, mi.Owner, mi.Beneficiary)
	switch contractAddr {
	case mi.Owner:
		if pondMiner.OwnerId == 0 {
			res.Status = contract.JobStatusResigning
		} else {
			res.Status = contract.JobStatusBeOn
			oldOwner, err := address.NewIDAddress(pondMiner.OwnerId)
			if err != nil {
				return res, err
			}
			owner, err = lotusrpc.FullApi.StateAccountKey(ctx, oldOwner, types.EmptyTSK)
			if err != nil {
				res.OldOwner = oldOwner.String()
			} else {
				res.OldOwner = owner.String()
			}
			res.OldOwnerId = fmt.Sprintf("f0%d", pondMiner.OwnerId)
		}
	case mi.Beneficiary:
		if pondMiner.OwnerId == 0 {
			res.Status = contract.JobStatusResigningNotBeneficiary
		} else {
			res.Status = contract.JobStatusBeOn
			res.OldOwner = res.Owner
			res.OldOwnerId = fmt.Sprintf("f0%d", pondMiner.OwnerId)
		}
	default:
		res.Status = contract.JobStatusDepart
		return res, nil
	}
	// Resignation order owner has not been changed and needs to be retrieved from the database
	if res.Status == contract.JobStatusResigning || res.Status == contract.JobStatusResigningNotBeneficiary {
		minerId, _ := strconv.ParseUint(actor, 10, 64)
		nodeInfo, err := new(NodeInfoService).GetSysNodeInfo(minerId)
		if err != nil {
			return res, err
		}
		res.OldOwnerId = nodeInfo.Owner
		oldOwner, err := address.NewFromString(nodeInfo.Owner)
		if err != nil {
			return res, err
		}
		owner, err = lotusrpc.FullApi.StateAccountKey(ctx, oldOwner, types.EmptyTSK)
		if err != nil {
			res.OldOwner = oldOwner.String()
		} else {
			res.OldOwner = owner.String()
		}
	}
	// Judging Debts
	if res.Status == contract.JobStatusBeOn {
		ethAddr, err := ethtypes.EthAddressFromFilecoinAddress(addr)
		if err != nil {
			return res, err
		}
		res.Debt, err = new(LiquidateService).GetDebt(ethAddr.String())
		if err != nil {
			return res, err
		}
	}
	return res, nil
}
