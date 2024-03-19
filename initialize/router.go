package initialize

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"zcfil-server/global"
	"zcfil-server/middleware"
	"zcfil-server/router"
)

// Initialize overall routing
func routers() *gin.Engine {
	Router := gin.Default()
	systemRouter := router.RouterGroupApp.System
	rpcRouter := router.RouterGroupApp.Rpc

	Router.Use(middleware.Cors()) // Directly release all cross domain requests
	global.ZC_LOG.Info("use middleware cors")
	Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	global.ZC_LOG.Info("register swagger handler")

	PublicGroup := Router.Group("")
	{
		// Health monitoring
		PublicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, "ok")
		})
	}

	systemRouter.InitSysHostGroupRouter(PublicGroup) // Node machine information management
	systemRouter.InitLiquidateRouter(PublicGroup)    // Liquidation
	systemRouter.InitMinerRouter(PublicGroup)        // Get node list
	systemRouter.InitNodeRecordRouter(PublicGroup)   // Get node operation records related
	systemRouter.InitLineChartRouter(PublicGroup)    // Obtain the Fil line chart of the total pool
	rpcRouter.InitLotusGroupRouter(PublicGroup)      // Lotus call

	return Router
}

// WebRouters Web routing
func WebRouters() *gin.Engine {
	Router := routers()
	return Router
}
