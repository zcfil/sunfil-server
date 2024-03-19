package cmd

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
	"zcfil-server/initialize"
	"zcfil-server/service"
)

var LiquidationCmd = &cli.Command{
	Name:  "liquidation",
	Usage: "Execute liquidation",
	Action: func(cctx *cli.Context) error {

		initialize.DBList()
		err := service.ServiceGroupApp.SystemServiceGroup.LiquidateService.Liquidation()
		if err != nil {
			log.Info("Liquidation err:", err)
			return err
		}

		return nil
	},
}
