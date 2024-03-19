package system

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/multisig"
	"github.com/filecoin-project/lotus/chain/consensus"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	cbor "github.com/ipfs/go-ipld-cbor"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"
	"log"
	"reflect"
	"sort"
	"strings"
	"zcfil-server/lotusrpc"
	"zcfil-server/model/rpc/request"
	"zcfil-server/model/rpc/response"
	"zcfil-server/utils"
)

// MsigInspectFunc Obtain signature information inspect
func (l *NodeQueryService) MsigInspectFunc(request *request.JsonRpc) *response.JsonRpcResult {
	if request == nil {
		return response.JsonRpcResultError(fmt.Errorf("param is empty"), 0)
	}
	log.Println("msigInspectFunc：", request.Params)
	if response.IsEmptyParams(request.Params) {
		return response.JsonRpcResultError(fmt.Errorf("param is empty"), 0)
	}

	ctx := context.Background()
	store := adt.WrapStore(ctx, cbor.NewCborStore(blockstore.NewAPIBlockstore(lotusrpc.FullApi)))
	if err := utils.VerifyParamType(request.Params.([]interface{})[0], reflect.String); err != nil {
		return response.JsonRpcResultError(err, request.Id)
	}
	maddr, err := address.NewFromString(request.Params.([]interface{})[0].(string))
	if err != nil {
		log.Println("1", err)
		return response.JsonRpcResultError(err, request.Id)
	}

	head, err := lotusrpc.FullApi.ChainHead(ctx)
	if err != nil {
		log.Println("12", err)
		return response.JsonRpcResultError(err, request.Id)
	}

	act, err := lotusrpc.FullApi.StateGetActor(ctx, maddr, head.Key())
	if err != nil {
		log.Println("13", err)
		return response.JsonRpcResultError(err, request.Id)
	}

	ownId, err := lotusrpc.FullApi.StateLookupID(ctx, maddr, types.EmptyTSK)
	if err != nil {
		log.Println("14", err)
		return response.JsonRpcResultError(err, request.Id)
	}
	log.Println(act.Code.String())
	mstate, err := multisig.Load(store, act)
	if err != nil {
		log.Println("15", err)
		return response.JsonRpcResultError(err, request.Id)
	}
	locked, err := mstate.LockedBalance(head.Height())
	if err != nil {
		log.Println("16", err)
		return response.JsonRpcResultError(err, request.Id)
	}
	var res response.MsigInfo
	res.Balance = act.Balance.String()
	res.Spendable = types.BigSub(act.Balance, locked).String()

	signers, err := mstate.Signers()
	if err != nil {
		log.Println("17", err)
		return response.JsonRpcResultError(err, request.Id)
	}
	threshold, err := mstate.Threshold()
	if err != nil {
		log.Println("18", err)
		return response.JsonRpcResultError(err, request.Id)
	}
	// By threshold
	res.Threshold = threshold

	// Can concurrently obtain
	for _, s := range signers {
		var addr response.AddressInfo
		addr.Id = s.String()
		signerActor, err := lotusrpc.FullApi.StateAccountKey(ctx, s, types.EmptyTSK)
		if err != nil {
			log.Println("19", err)
			return response.JsonRpcResultError(err, request.Id)
		}
		addr.Address = signerActor.String()
		nonce, err := lotusrpc.FullApi.MpoolGetNonce(ctx, signerActor)
		if err != nil {
			log.Println("20", err)
			return response.JsonRpcResultError(err, request.Id)
		}
		addr.Nonce = nonce
		res.Signers = append(res.Signers, addr)
	}

	pending := make(map[int64]multisig.Transaction)
	if err := mstate.ForEachPendingTxn(func(id int64, txn multisig.Transaction) error {
		pending[id] = txn
		return nil
	}); err != nil {
		log.Println("21", err)
		return response.JsonRpcResultError(err, request.Id)
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
		var methodMeta vm.MethodMeta
		for _, vmap := range methods.Methods {
			for k, v := range vmap {
				if k == 23 {
					methodMeta = v
				}
			}
		}
		for _, txid := range txids {
			tx := pending[txid]
			target := tx.To.String()
			if tx.To == ownId {
				target += " (self)"
			}
			targAct, err := lotusrpc.FullApi.StateGetActor(ctx, tx.To, types.EmptyTSK)
			paramStr := fmt.Sprintf("%x", tx.Params)
			if err != nil {
				if tx.Method == 0 {
					fmt.Printf("%d\t%s\t%d\t%s\t%s\t%s(%d)\t%s\n", txid, "pending", len(tx.Approved), target, types.FIL(tx.Value), "Send", tx.Method, paramStr)
				} else {
					fmt.Printf("%d\t%s\t%d\t%s\t%s\t%s(%d)\t%s\n", txid, "pending", len(tx.Approved), target, types.FIL(tx.Value), "new account, unknown method", tx.Method, paramStr)
				}
			} else {
				fmt.Println(methodMeta, targAct.Code, tx.Method)
				if tx.Method != 0 && methodMeta.Params != nil {
					ptyp := reflect.New(methodMeta.Params.Elem()).Interface().(cbg.CBORUnmarshaler)
					if err := ptyp.UnmarshalCBOR(bytes.NewReader(tx.Params)); err != nil {
						log.Println("22", err)
						return response.JsonRpcResultError(xerrors.Errorf("failed to decode parameters of transaction %d: %w", txid, err), request.Id)
					}

					b, err := json.Marshal(ptyp)
					if err != nil {
						log.Println("23", err)
						return response.JsonRpcResultError(xerrors.Errorf("could not json marshal parameter type: %w", err), request.Id)
					}

					paramStr = strings.ReplaceAll(string(b), `"`, "")
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
	return response.JsonRpcResultOk(res, request.Id)
}
