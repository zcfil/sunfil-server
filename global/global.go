package global

import (
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/go-redis/redis/v8"
	"sync"

	"zcfil-server/utils/timer"

	"golang.org/x/sync/singleflight"

	"go.uber.org/zap"

	"zcfil-server/config"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	ZC_DB     *gorm.DB
	ZC_DBList map[string]*gorm.DB
	ZC_REDIS  *redis.Client
	ZC_CONFIG config.Server
	ZC_VP     *viper.Viper

	ZC_LOG                 *zap.Logger
	ZC_Timer               timer.Timer = timer.NewTimerTask()
	ZC_Concurrency_Control             = &singleflight.Group{}
	PushNoticeChan                     = make(chan *types.Message, 128)
	lock                   sync.RWMutex
)

// GetGlobalDBByDBName Obtain the db in the db list by name
func GetGlobalDBByDBName(dbname string) *gorm.DB {
	lock.RLock()
	defer lock.RUnlock()
	return ZC_DBList[dbname]
}

// MustGetGlobalDBByDBName Obtain db by name. If it does not exist, panic
func MustGetGlobalDBByDBName(dbname string) *gorm.DB {
	lock.RLock()
	defer lock.RUnlock()
	db, ok := ZC_DBList[dbname]
	if !ok || db == nil {
		panic("db no init")
	}
	return db
}
