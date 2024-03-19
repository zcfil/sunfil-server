package core

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
	"zcfil-server/global"
	"zcfil-server/initialize"
)

type server interface {
	ListenAndServe() error
}

func RunWebServer() {
	// Initialize Redis service
	//initialize.Redis()

	Router := initialize.WebRouters()
	Router.Static("/form-generator", "./resource/page")

	address := fmt.Sprintf(":%d", global.ZC_CONFIG.System.Addr)
	s := initServerNew(address, Router)

	time.Sleep(10 * time.Microsecond)
	global.ZC_LOG.Info("server run success on ", zap.String("address", address))

	global.ZC_LOG.Error(s.ListenAndServe().Error())
}

func initServerNew(address string, router *gin.Engine) server {
	return &http.Server{
		Addr:           address,
		Handler:        router,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}
