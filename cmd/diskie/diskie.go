package main

import (
	"diskie"
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
		Name:  "diskie",
		Usage: "Command line tool for UDisks2",
		Commands: []cli.Command{
			{
				Name:  "blockdevs",
				Usage: "Print block devices",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "format",
						Value: "json",
						Usage: `Output format. Can be "json" or a golang template.`,
					},
					&cli.UintFlag{
						Name:  "min-importance",
						Value: 0,
						Usage: "Only include block devices more important than the given level. Possible values are 0 through 3.",
					},
				},
				Action: func(c *cli.Context) error {
					f := c.String("format")
					i := c.Uint("min-importance")
					if i > 3 {
						return fmt.Errorf("min-importance of %d is out of the possible range of 0 through 3", i)
					}
					if f == "json" {
						return blockdevs(i)
					}
					return fmt.Errorf(`the only supported format is "json"`)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func blockdevs(importance uint) error {
	dsk, err := diskie.Connect()
	if err != nil {
		return fmt.Errorf("could not create diskie client: %w", err)
	}

	blockmap, err := dsk.BlockDevices()
	if err != nil {
		return fmt.Errorf("could not get block devices: %w", err)
	}

	blocks := blockmap.Sort()
	blocks = blockmap.Filter(blocks, importance)

	return pretty(blocks)
}

func pretty(obj interface{}) error {
	output, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		return fmt.Errorf("could not marshal object into json: %w", err)
	}
	fmt.Println(string(output))
	return nil
}
