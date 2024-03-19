package rpc

import (
	"encoding/json"
	"fmt"
	"log"
	"zcfil-server/global"
	"zcfil-server/model/rpc/request"
	"zcfil-server/model/rpc/response"
)

type LotusMpoolService struct{}

func (l *LotusMpoolService) MpoolPush(req *request.JsonRpc) *response.JsonRpcResult {
	if req == nil {
		return response.JsonRpcResultError(fmt.Errorf("param is empty"), 0)
	}
	res := LotusV1Service{}.RequestLotusFunc(req)
	if res.Error != nil {
		return res
	}
	param, err := json.Marshal(req.Params)
	if err != nil {
		log.Println("error1:", err.Error())
		return res
	}
	var msg []request.JsonParam
	if err = json.Unmarshal(param, &msg); err != nil {
		log.Println("error2:", err.Error())
		return res
	}
	go func() {
		if len(msg) > 0 {
			log.Println("Trigger push：", msg)
			global.PushNoticeChan <- &msg[0].Message
		}
	}()

	return res
}
