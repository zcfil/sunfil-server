package core

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"path/filepath"
	"zcfil-server/core/internal"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"

	"zcfil-server/global"
	_ "zcfil-server/packfile"
)

func Viper(path ...string) *viper.Viper {
	var config string

	if len(path) == 0 {
		flag.StringVar(&config, "c", "", "choose config file.")
		flag.Parse()
		if config == "" {
			if configEnv := os.Getenv(internal.ConfigEnv); configEnv == "" { //  Determine whether the environment variable stored in constant internal.ConfigEnv is empty
				switch gin.Mode() {
				case gin.DebugMode:
					config = internal.ConfigDefaultFile
					fmt.Printf("You are using the %s environment name in gin mode, and the path to config is %s\n", gin.EnvGinMode, internal.ConfigDefaultFile)
				case gin.ReleaseMode:
					config = internal.ConfigReleaseFile
					fmt.Printf("You are using the %s environment name in gin mode, and the path to config is %s\n", gin.EnvGinMode, internal.ConfigReleaseFile)
				case gin.TestMode:
					config = internal.ConfigTestFile
					fmt.Printf("You are using the %s environment name in gin mode, and the path to config is %s\n", gin.EnvGinMode, internal.ConfigTestFile)
				}
			} else { // The environment variable stored in constant internal.ConfigEnv is not empty, and the value is assigned to config
				config = configEnv
				fmt.Printf("您正在使用%s环境变量,config的路径为%s\n", internal.ConfigEnv, config)
			}
		} else { // Command line parameter is not empty, assign value to config
			fmt.Printf("You are using the value passed by the - c parameter on the command line, and the path to config is %s\n", config)
		}
	} else { // The first value of the variable parameter passed by the function is assigned to config
		config = path[0]
		fmt.Printf("You are using the value passed by Func Viper(), and the path to config is %s\n", config)
	}

	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err = v.Unmarshal(&global.ZC_CONFIG); err != nil {
			fmt.Println(err)
		}
	})
	if err = v.Unmarshal(&global.ZC_CONFIG); err != nil {
		fmt.Println(err)
	}

	global.ZC_CONFIG.AutoCode.Root, _ = filepath.Abs("..")
	return v
}
