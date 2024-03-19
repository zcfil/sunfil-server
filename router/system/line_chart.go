package system

import (
	"github.com/gin-gonic/gin"
	v1 "zcfil-server/api/v1"
)

type LineChartRouter struct{}

func (s *LineChartRouter) InitLineChartRouter(Router *gin.RouterGroup) {
	lineChartRouter := Router.Group("lineChart")
	// Pledge related
	stakeInfoApi := v1.ApiGroupApp.SystemApiGroup.StakeInfoApi
	{
		lineChartRouter.GET("getTotalFilList", stakeInfoApi.GetTotalFilList)
		lineChartRouter.GET("getStakeInfoList", stakeInfoApi.GetStakeInfoList)
		lineChartRouter.GET("getTotalityChange", stakeInfoApi.GetTotalityChangeList)

		lineChartRouter.GET("getTotalFilSide", stakeInfoApi.GetTotalFilSide)
		lineChartRouter.GET("getTotalRateSide", stakeInfoApi.GetTotalRateSide)
		lineChartRouter.GET("getLoanRateSide", stakeInfoApi.GetLoanRateSide)
		lineChartRouter.GET("getStackRateSide", stakeInfoApi.GetStackRateSide)

		lineChartRouter.GET("getFinanceUseRate", stakeInfoApi.GetFinanceUseRate)
	}

	// Lending related
	debtInfoApi := v1.ApiGroupApp.SystemApiGroup.DebtInfoApi
	{
		lineChartRouter.GET("getDebtInfoList", debtInfoApi.GetDebtInfoList)
	}
}
