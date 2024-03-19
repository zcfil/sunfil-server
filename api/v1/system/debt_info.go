package system

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"math"
	"time"
	"zcfil-server/global"
	"zcfil-server/model/common/request"
	"zcfil-server/model/common/response"
	systemResponse "zcfil-server/model/system/response"
)

type DebtInfoApi struct{}

type DebtInfoResp struct {
	RecordTime  time.Time `json:"recordTime"`
	DebtBalance float64   `json:"debtBalance"`
	DebtRate    float64   `json:"debtRate"`
}

// GetDebtInfoList Obtain loan information
func (s *DebtInfoApi) GetDebtInfoList(c *gin.Context) {
	var req request.GetFoldLineListReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, err := debtInfoService.GetSysDebtInfoList(req)
	if err != nil {
		global.ZC_LOG.Error("Getting information failure!", zap.Error(err))
		response.FailWithMessage("Getting information failure", c)
		return
	}

	totalFilResp := make([]DebtInfoResp, len(list))
	for key, val := range list {
		_createdAt := val.CreatedAt.Format("2006-01-02 15:00:00")
		dealTime, _ := time.ParseInLocation("2006-01-02 15:04:05", _createdAt, time.Local)
		totalFilResp[key].RecordTime = dealTime
		totalFilResp[key].DebtBalance = val.DebtBalance
		totalFilResp[key].DebtRate = val.DebtRate
	}

	baseRate := math.Pow(10, 18)
	sideResp := &systemResponse.GetLoanRateSideResp{}
	dealLoanRateSideResp(sideResp, baseRate)
	currentTotalResp := DebtInfoResp{
		RecordTime:  time.Now(),
		DebtBalance: sideResp.TotalLoan,
		DebtRate:    sideResp.LoanAPY,
	}

	resp := make([]DebtInfoResp, 0)
	resp = append(resp, currentTotalResp)
	resp = append(resp, totalFilResp...)

	response.OkWithDetailed(response.PageResult{
		List: resp,
	}, "Success", c)
}
