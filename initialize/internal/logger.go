package internal

import (
	"fmt"

	"gorm.io/gorm/logger"
	"zcfil-server/global"
)

type writer struct {
	logger.Writer
}

func NewWriter(w logger.Writer) *writer {
	return &writer{Writer: w}
}

func (w *writer) Printf(message string, data ...interface{}) {
	var logZap bool
	switch global.ZC_CONFIG.System.DbType {
	case "mysql":
		logZap = global.ZC_CONFIG.Mysql.LogZap
	case "pgsql":
		logZap = global.ZC_CONFIG.Pgsql.LogZap
	}
	if logZap {
		global.ZC_LOG.Info(fmt.Sprintf(message+"\n", data...))
	} else {
		w.Writer.Printf(message, data...)
	}
}
