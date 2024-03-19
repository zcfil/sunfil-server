package contract

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	eabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	builtintypes "github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"
	"log"
	"strings"
	"zcfil-server/lotusrpc"
)

const (
	DefaultWallet        = "0x1d8f134e0e3D45c832bB3cdAB51b723F5CCF59C2"
	DefaultControlWallet = "f3qlu47iqiuosvfiu5tfck26acqy3wy2vvrf4sfypca4ax7qbvmq4bf5dxjtkahv6b7dppocq34migysrspooq"
)

type ResPushInfo struct {
	MsgId string
}
type Contract struct {
	Addr     ContractAddr
	PushAddr ContractAddr
	MyABI    eabi.ABI
}
type ContractAddr string

func (ca ContractAddr) ToFilecionAddress() address.Address {
	faddr, err := address.NewFromString(string(ca))
	if err != nil { // This isn't a filecoin address
		eaddr, err := ethtypes.ParseEthAddress(string(ca))
		if err != nil { // This isn't an Eth address either
			log.Fatal("address is not a filecoin or eth address", err, faddr)
			return faddr
		}
		faddr, err = eaddr.ToFilecoinAddress()
		if err != nil {
			log.Fatal("address is not a filecoin or eth address", err, faddr)
			return faddr
		}
	}
	return faddr
}

func (ca ContractAddr) ToActorID() address.Address {
	faddr, err := address.NewFromString(string(ca))
	if err != nil { // This isn't a filecoin address
		eaddr, err := ethtypes.ParseEthAddress(string(ca))
		if err != nil { // This isn't an Eth address either
			log.Fatal("address is not a filecoin or eth address", err, faddr)
			return faddr
		}
		faddr, err = eaddr.ToFilecoinAddress()
		if err != nil {
			log.Fatal("address is not a filecoin or eth address", err, faddr)
			return faddr
		}
	}
	if faddr.Protocol() != address.ID {
		faddr, err = lotusrpc.FullApi.StateLookupID(context.Background(), faddr, types.EmptyTSK)
		if err != nil {
			return faddr
		}
	}
	return faddr
}

func (ca ContractAddr) ToEthAddress() ethtypes.EthAddress {
	eaddr, err := ethtypes.ParseEthAddress(string(ca))
	if err != nil { // This isn't a filecoin address
		faddr, err := address.NewFromString(string(ca))
		if err != nil { // This isn't an Eth address either
			log.Fatal("address is not a filecoin or eth address", err, faddr)
			return eaddr
		}
		eaddr, _, err = ethAddrFromFilecoinAddress(context.Background(), faddr)
		if err != nil {
			log.Fatal("address is not a filecoin or eth address", err, faddr)
			return eaddr
		}
	}
	return eaddr
}

func (ca ContractAddr) ToString() string {
	return string(ca)
}

func ethAddrFromFilecoinAddress(ctx context.Context, addr address.Address) (ethtypes.EthAddress, address.Address, error) {
	var faddr address.Address
	var err error

	switch addr.Protocol() {
	case address.BLS, address.SECP256K1:
		faddr, err = lotusrpc.FullApi.StateLookupID(ctx, addr, types.EmptyTSK)
		if err != nil {
			return ethtypes.EthAddress{}, addr, err
		}
	case address.Actor, address.ID:
		faddr, err = lotusrpc.FullApi.StateLookupID(ctx, addr, types.EmptyTSK)
		if err != nil {
			return ethtypes.EthAddress{}, addr, err
		}
		fAct, err := lotusrpc.FullApi.StateGetActor(ctx, faddr, types.EmptyTSK)
		if err != nil {
			return ethtypes.EthAddress{}, addr, err
		}
		if fAct.Address != nil && (*fAct.Address).Protocol() == address.Delegated {
			faddr = *fAct.Address
		}
	case address.Delegated:
		faddr = addr
	default:
		return ethtypes.EthAddress{}, addr, xerrors.Errorf("Filecoin address doesn't match known protocols")
	}

	ethAddr, err := ethtypes.EthAddressFromFilecoinAddress(faddr)
	if err != nil {
		return ethtypes.EthAddress{}, addr, err
	}

	return ethAddr, faddr, nil
}
func (ca *Contract) SetABI(contractAbi string) {
	myabi, err := eabi.JSON(strings.NewReader(contractAbi))
	if err != nil {
		log.Fatal(err)
	}
	ca.MyABI = myabi
}

func (c *Contract) PushContract(param []byte) (ResPushInfo, error) {

	var res ResPushInfo
	calldata, err := ethtypes.DecodeHexStringTrimSpace(hexutil.Encode(param))
	if err != nil {
		log.Println("decoding hex input data: %w", err)
		return res, err
	}

	var buffer bytes.Buffer
	if err := cbg.WriteByteArray(&buffer, calldata); err != nil {
		log.Println("failed to encode evm params as cbor: %w", err)
		return res, err
	}

	addr, err := address.NewFromString(c.PushAddr.ToString())
	if err != nil {
		return res, xerrors.Errorf("failed to decode address: %w", err)
	}

	eaddr, err := ethtypes.ParseEthAddress(DefaultWallet)
	if err != nil {
		return res, xerrors.Errorf("address is not a filecoin or eth address")
	}
	faddr, err := eaddr.ToFilecoinAddress()
	if err != nil {
		return res, err
	}

	calldata = buffer.Bytes()
	msg := &types.Message{
		To:     addr,
		From:   faddr,
		Value:  abi.NewTokenAmount(0),
		Method: builtintypes.MethodsEVM.InvokeContract,
		Params: calldata,
	}

	log.Println("sending message...")
	log.Println(fmt.Sprintf("msg:%+v", msg))
	smsg, err := lotusrpc.FullApi.MpoolPushMessage(context.Background(), msg, nil)
	if err != nil {
		log.Println("failed to push message: %w", err)
		return res, err
	}
	log.Println("waiting for message to execute...", smsg.Cid())
	wait, err := lotusrpc.FullApi.StateWaitMsg(context.Background(), smsg.Cid(), 0, 0, false)
	if err != nil {
		log.Println("error waiting for message: %w", err)
		return res, err
	}
	log.Println("gas used", wait.Receipt.GasUsed)
	// check it executed successfully
	if wait.Receipt.ExitCode != 0 {
		return res, xerrors.Errorf("exitCode is err:", wait.Receipt.ExitCode)
	} else {
		res.MsgId = smsg.Cid().String()
	}

	log.Println("OK")
	return res, nil
}

func (c *Contract) CallContract(param []byte, res ...interface{}) error {
	methodId := param[:4]
	callData, err := ethtypes.DecodeHexStringTrimSpace(hexutil.Encode(param))
	if err != nil {
		log.Println("decoding hex input data: %w", err)
		return err
	}

	fromEthAddr, err := ethtypes.ParseEthAddress(DefaultWallet)
	if err != nil {
		return err
	}

	toEthAddr := c.Addr.ToEthAddress()

	resData, err := lotusrpc.FullApi.EthCall(context.Background(), ethtypes.EthCall{
		From: &fromEthAddr,
		To:   &toEthAddr,
		Data: callData,
	}, "latest")
	if err != nil {
		fmt.Println("Eth call fails, return val: ", err)
		return err
	}

	return c.DecodeValue(hexutil.Encode(methodId), hex.EncodeToString(resData), res...)
}

func (c *Contract) DecodeValue(methodData, txData string, res ...interface{}) error {
	if strings.HasPrefix(methodData, "0x") {
		methodData = methodData[2:]
	}

	if strings.HasPrefix(txData, "0x") {
		txData = txData[2:]
	}

	decodedSig, err := hex.DecodeString(methodData)
	if err != nil {
		return err
	}

	method, err := c.MyABI.MethodById(decodedSig)
	if err != nil {
		return err
	}

	decodedData, err := hex.DecodeString(txData)
	if err != nil {
		return err
	}

	Outputs, err := method.Outputs.Unpack(decodedData)
	if err != nil {
		return err
	}

	for i, output := range Outputs {
		if len(res) <= i {
			break
		}

		buf, err := json.Marshal(output)
		if err != nil {
			return err
		}

		if err = json.Unmarshal(buf, &res[i]); err != nil {
			if res[i] != nil {
				return nil
			}
			return err
		}
	}

	return nil
}

func (c *Contract) DecodeInput(txData string, res ...interface{}) error {
	if strings.HasPrefix(txData, "0x") {
		txData = txData[2:]
	}

	decodedSig, err := hex.DecodeString(txData[:8])
	if err != nil {
		return err
	}

	method, err := c.MyABI.MethodById(decodedSig)
	if err != nil {
		return err
	}

	decodedData, err := hex.DecodeString(txData[8:])
	if err != nil {
		return err
	}

	inputs, err := method.Inputs.Unpack(decodedData)
	if err != nil {
		return err
	}
	for i, v := range inputs {
		if i == len(res) {
			return nil
		}
		s, _ := json.Marshal(v)
		json.Unmarshal(s, res[i])
	}
	return nil
}
