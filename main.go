package main

import (
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"log"
	"os"
	"zcfil-server/build"
	"zcfil-server/cmd"
	"zcfil-server/core"
	"zcfil-server/global"
	"zcfil-server/initialize"
)

//go:generate go env -w GO111MODULE=on
//go:generate go env -w GOPROXY=https://goproxy.cn,direct
//go:generate go mod tidy
//go:generate go mod download
func init() {
	build.InitContracts()
}
func main() {
	local := []*cli.Command{
		cmd.Run,
		cmd.SetHeight,
		cmd.GetHeight,
		cmd.LiquidationCmd,
	}
	app := &cli.App{
		Name:                 "zcfil-server",
		Usage:                "zcfil-server",
		EnableBashCompletion: true,
		Commands:             local,
	}
	//读取配置
	global.ZC_VP = core.Viper("config/config.yaml")
	global.ZC_LOG = core.Zap()
	zap.ReplaceGlobals(global.ZC_LOG)
	global.ZC_DB = initialize.Gorm()
	err := app.Run(os.Args)
	if err != nil {
		log.Println("Start failed:", err)
	}
}
