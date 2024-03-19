package request

import (
	"fmt"

	"zcfil-server/config"
)

type InitDB struct {
	DBType   string `json:"dbType"`                      // Database type
	Host     string `json:"host"`                        // Server address
	Port     string `json:"port"`                        // Database connection port
	UserName string `json:"userName" binding:"required"` // Database username
	Password string `json:"password"`                    // Database password
	DBName   string `json:"dbName" binding:"required"`   // Database name
}

// MysqlEmptyDsn Msyql empty database construction link
func (i *InitDB) MysqlEmptyDsn() string {
	if i.Host == "" {
		i.Host = "127.0.0.1"
	}
	if i.Port == "" {
		i.Port = "3306"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/", i.UserName, i.Password, i.Host, i.Port)
}

// ToMysqlConfig Conversion config.Mysql
func (i *InitDB) ToMysqlConfig() config.Mysql {
	return config.Mysql{
		GeneralDB: config.GeneralDB{
			Path:         i.Host,
			Port:         i.Port,
			Dbname:       i.DBName,
			Username:     i.UserName,
			Password:     i.Password,
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			LogMode:      "error",
			Config:       "charset=utf8mb4&parseTime=True&loc=Local",
		},
	}
}
