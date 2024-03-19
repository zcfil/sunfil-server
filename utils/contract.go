package utils

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	builtintypes "github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/lotus/api/v1api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	cbg "github.com/whyrusleeping/cbor-gen"
	"log"
	"zcfil-server/lotusrpc"
	m "zcfil-server/model/system"
)

func InvokeContractMethod(api v1api.FullNode, contractAbi m.SysContractAbi) (OutputData, error) {

	var outData OutputData
	var fromAddr address.Address
	var toAddr address.Address
	toAddr, err := address.NewFromString(contractAbi.ActorID)
	if err != nil {
		return outData, err
	}
	fromAddr, err = address.NewFromString(contractAbi.Wallet)
	if err != nil {
		return outData, err
	}

	methodId := crypto.Keccak256(contractAbi.Param)[:4]

	var calldata []byte
	calldata, err = ethtypes.DecodeHexStringTrimSpace(hexutil.Encode(methodId))
	if err != nil {
		log.Println("decoding hex input data: %w", err)
		return outData, err
	}

	var buffer bytes.Buffer
	if err := cbg.WriteByteArray(&buffer, calldata); err != nil {
		log.Println("failed to encode evm params as cbor: %w", err)
		return outData, err
	}

	calldata = buffer.Bytes()
	msg := &types.Message{
		To:     toAddr,
		From:   fromAddr,
		Value:  abi.NewTokenAmount(0),
		Method: builtintypes.MethodsEVM.InvokeContract,
		Params: calldata,
	}

	log.Println("sending message...")
	smsg, err := api.MpoolPushMessage(context.Background(), msg, nil)
	if err != nil {
		log.Println("failed to push message: %w", err)
		return outData, err
	}
	log.Println("waiting for message to execute...")
	wait, err := api.StateWaitMsg(context.Background(), smsg.Cid(), 0, 0, false)
	if err != nil {
		log.Println("error waiting for message: %w", err)
		return outData, err
	}
	// check it executed successfully
	if wait.Receipt.ExitCode != 0 {
		log.Println("actor execution failed")
		return outData, err
	}

	log.Println("Gas used: ", wait.Receipt.GasUsed)
	result, err := cbg.ReadByteArray(bytes.NewBuffer(wait.Receipt.Return), uint64(len(wait.Receipt.Return)))
	if err != nil {
		log.Println("evm result not correctly encoded: %w", err)
		return outData, err
	}

	txDataDecoder := NewABIDecoder()
	txDataDecoder.SetABI(contractAbi.AbiContent)
	method, err := txDataDecoder.DecodeOutPut(hexutil.Encode(methodId), hex.EncodeToString(result))
	if err != nil {
		log.Fatal(err)
	}
	outData = method

	return outData, nil
}

func CallContractMethod(contractAbi m.SysContractAbi) (OutputData, error) {

	fmt.Println(fmt.Sprintf("CallContractMethod param:%+v", contractAbi))
	var outData OutputData
	var data []byte
	methodStr := []byte(contractAbi.AbiName)
	methodId := crypto.Keccak256(methodStr)[:4]
	data = append(data, methodId...)
	data = append(data, contractAbi.Param...)

	callData, err := ethtypes.DecodeHexStringTrimSpace(hexutil.Encode(data))
	if err != nil {
		log.Println("decoding hex input data: %w", err)
		return outData, err
	}

	fromEthAddr, err := ethtypes.ParseEthAddress(contractAbi.Wallet)
	if err != nil {
		return outData, err
	}

	toEthAddr, err := ethtypes.ParseEthAddress(contractAbi.ActorID)
	if err != nil {
		return outData, err
	}

	res, err := lotusrpc.FullApi.EthCall(context.Background(), ethtypes.EthCall{
		From: &fromEthAddr,
		To:   &toEthAddr,
		Data: callData,
	}, "latest")
	if err != nil {
		fmt.Println("Eth call fails, return val: ", res)
		return outData, err
	}

	txDataDecoder := NewABIDecoder()
	txDataDecoder.SetABI(contractAbi.AbiContent)
	method, err := txDataDecoder.DecodeOutPut(hexutil.Encode(methodId), hex.EncodeToString(res))
	if err != nil {
		log.Fatal(err)
	}
	outData = method

	return outData, nil
}
