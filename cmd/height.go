package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"strconv"
	"zcfil-server/service"
)

var SetHeight = &cli.Command{
	Name:  "set-height",
	Usage: "Set listening height",
	Action: func(cctx *cli.Context) error {
		height, err := strconv.ParseInt(cctx.Args().First(), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		if err = service.ServiceGroupApp.SystemServiceGroup.UpdateHeight(height); err != nil {
			log.Fatal(err)
		}
		return err
	},
}

var GetHeight = &cli.Command{
	Name:  "height",
	Usage: "Obtain listening height",
	Flags: []cli.Flag{},
	Action: func(cctx *cli.Context) error {
		height, err := service.ServiceGroupApp.SystemServiceGroup.GetHeight()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(height)
		return nil
	},
}
