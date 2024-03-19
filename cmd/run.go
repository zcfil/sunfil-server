package cmd

import (
	"github.com/urfave/cli/v2"
	"log"
	"zcfil-server/core"
	"zcfil-server/global"
	"zcfil-server/initialize"
	"zcfil-server/lotusrpc"
	"zcfil-server/service"
)

var Run = &cli.Command{
	Name:  "run",
	Usage: "Running the system",
	Flags: []cli.Flag{},
	Action: func(cctx *cli.Context) error {
		initialize.DBList()
		if global.ZC_DB != nil {
			initialize.RegisterTables(global.ZC_DB) // 初始化表
			log.Println(initialize.InitMysqlData())
			db, _ := global.ZC_DB.DB()
			defer db.Close()
		}
		//Connecting Lotus Api1
		if _, _, err := lotusrpc.NewLotusApi(); err != nil {
			log.Fatal("lotus connect fail!")
		}
		//Connecting Lotus Api0
		if _, _, err := lotusrpc.NewLotusApi0(); err != nil {
			log.Fatal("lotus connect fail!")
		}
		if err := service.ServiceGroupApp.SystemServiceGroup.InitHeight(); err != nil {
			log.Fatal("InitHeight fail!")
		}

		//Timed acquisition of alarm nodes
		go initialize.ContractWarnNode()
		// Timed deduction
		go initialize.TimingRepayment()

		// Timed tasks, timed pull of pledge pool fil changes, pledge interest rate, and total pledge amount
		go initialize.RecordStackPoolBalance()
		// Timed tasks, pulling loan interest rates and total loan amounts at regular intervals
		go initialize.RecordDebtInfo()
		// Timed tasks, timed pulling of node loan limits
		go initialize.RecordNodeDebtNum()

		// Scan the chain to obtain operation records
		go initialize.ChainTicker()

		// Synchronize pledge information to the contract
		go initialize.SynPledgeInfo()

		//Update pledge information
		go initialize.SaveBlockRewards()

		// Regularly update database node operator information, maximum debt ratio, and liquidation threshold
		go initialize.UpdateNodeOperation()

		// Regularly updating cache information
		go initialize.UpdateSolidityData()

		core.RunWebServer()
		return nil
	},
}
