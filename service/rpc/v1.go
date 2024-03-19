package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
	"zcfil-server/global"
	"zcfil-server/model/rpc/request"
	"zcfil-server/model/rpc/response"
)

type LotusV1Service struct{}

// HTTP
func (l LotusV1Service) RequestLotusFunc(request *request.JsonRpc) *response.JsonRpcResult {
	client := &http.Client{
		Timeout: time.Minute * 5,
	}
	var buf []byte
	if request == nil {
		return response.JsonRpcResultError(fmt.Errorf("param is empty"), 0)
	}
	var err error
	buf, err = json.Marshal(request)
	if err != nil {
		log.Println("Marshal error：", err.Error())
		return response.JsonRpcResultError(err, request.Id)
	}

	path := "http://" + global.ZC_CONFIG.Lotus.Host + "/rpc/v1"
	req, err := http.NewRequest("POST", path, bytes.NewReader(buf))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return response.JsonRpcResultError(err, request.Id)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", strconv.FormatInt(req.ContentLength, 10))
	if global.ZC_CONFIG.Lotus.Token != "" {
		req.Header.Set("Authorization", "Bearer "+global.ZC_CONFIG.Lotus.Token)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("do error：", err.Error())
		return response.JsonRpcResultError(err, request.Id)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var res response.JsonRpcResult
	if len(data) > 0 {
		_ = json.Unmarshal(data, &res)
	}

	return &res
}
