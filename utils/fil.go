package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-address"
	builtintypes "github.com/filecoin-project/go-state-types/builtin"
	"github.com/filecoin-project/lotus/chain/consensus"
	"github.com/filecoin-project/lotus/chain/types/ethtypes"
	cbg "github.com/whyrusleeping/cbor-gen"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	AttoFIL = 18
	NanoFIL = 9
)

func NanoOrAttoToFIL(fil string, filtype int) (res float64, err error) {
	// Greater than 18 or 9 bits
	if len(fil) > filtype {
		str := fil[0:len(fil)-filtype] + "." + fil[len(fil)-filtype:]
		res, err = strconv.ParseFloat(str, 64)
		return
	}
	// Less than 18 or 9 bits
	str := "0."
	for i := 0; i < filtype-len(fil); i++ {
		str += "0"
	}
	str = str + fil
	res, err = strconv.ParseFloat(str, 64)
	return
}

func BlockHeight() int64 {
	//dataStr := "2022-11-02 02:13:00" // Calibration network reset time
	dataStr := "2020-08-25 06:00:00" // Official website
	t, _ := time.ParseInLocation("2006-01-02 15:04:05 ", dataStr, time.Local)
	t1 := time.Now().UnixNano() / 1e6
	t2 := t.UnixNano() / 1e6
	num := (t1 - t2) / 30 / 1000

	return int64(num)
}

func BlockHeightToTime(num int64) time.Time {
	num = num * 30 * 1e3
	dataStr := "2020-08-25 06:00:00"
	t, _ := time.ParseInLocation("2006-01-02 15:04:05 ", dataStr, time.Local)
	t2 := t.UnixNano() / 1e6
	t1 := (t2 + num) / 1e3

	return time.Unix(t1, 0)
}

func AnalyseParams(params []byte) string {
	methods := consensus.NewActorRegistry()
	var paramStr string
	for _, v := range methods.Methods {
		method := v[builtintypes.MethodsEVM.InvokeContract]
		if method.Params != nil {
			ptyp := reflect.New(method.Params.Elem()).Interface().(cbg.CBORUnmarshaler)
			if err := ptyp.UnmarshalCBOR(bytes.NewReader(params)); err != nil {
				//log.Println("failed to decode parameters of transaction ", txid, err)
				continue
			}

			b, err := json.Marshal(ptyp)
			if err != nil {
				fmt.Println("could not json marshal parameter type: %w", err, ptyp)
				continue
			}
			if string(b) == "{}" {
				continue
			}
			paramStr = strings.ReplaceAll(string(b), `"`, "")
			break
		}
	}
	return paramStr
}

func GetAddr(walletAddr string) address.Address {

	faddr, err := address.NewFromString(string(walletAddr))
	if err != nil { // This isn't a filecoin address
		eaddr, err := ethtypes.ParseEthAddress(string(walletAddr))
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
