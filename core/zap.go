package core

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"zcfil-server/core/internal"
	"zcfil-server/global"
	"zcfil-server/utils"
)

func Zap() (logger *zap.Logger) {
	if ok, _ := utils.PathExists(global.ZC_CONFIG.Zap.Director); !ok {
		fmt.Printf("create %v directory\n", global.ZC_CONFIG.Zap.Director)
		_ = os.Mkdir(global.ZC_CONFIG.Zap.Director, os.ModePerm)
	}

	cores := internal.Zap.GetZapCores()
	logger = zap.New(zapcore.NewTee(cores...))

	if global.ZC_CONFIG.Zap.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}
	return logger
}
