package lotusrpc

import (
	"context"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/api/v0api"

	"github.com/filecoin-project/lotus/api/v1api"
	"log"
	"net/http"
	"zcfil-server/global"
)

var FullApi v1api.FullNode
var FullApi0 v0api.FullNode

// NewLotusApi Connecting Lotus
func NewLotusApi() (v1api.FullNode, jsonrpc.ClientCloser, error) {
	headers := http.Header{}
	if global.ZC_CONFIG.Lotus.Token != "" {
		headers.Add("Authorization", "Bearer "+global.ZC_CONFIG.Lotus.Token)
	}
	var cl jsonrpc.ClientCloser
	var err error
	//var err error
	FullApi, cl, err = client.NewFullNodeRPCV1(context.Background(), "http://"+global.ZC_CONFIG.Lotus.Host+"/rpc/v1", headers)
	if err != nil {
		log.Println("connecting with lotus failed: ", err.Error())
		return nil, nil, err
	}
	return FullApi, cl, nil
}

// NewLotusApi0 Connecting Lotus
func NewLotusApi0() (v0api.FullNode, jsonrpc.ClientCloser, error) {
	headers := http.Header{}
	if global.ZC_CONFIG.Lotus.Token != "" {
		headers.Add("Authorization", "Bearer "+global.ZC_CONFIG.Lotus.Token)
	}
	var cl jsonrpc.ClientCloser
	var err error
	//var err error
	FullApi0, cl, err = client.NewFullNodeRPCV0(context.Background(), "http://"+global.ZC_CONFIG.Lotus.Host+"/rpc/v1", headers)
	if err != nil {
		log.Println("connecting with lotus failed: ", err.Error())
		return nil, nil, err
	}
	return FullApi0, cl, nil
}
