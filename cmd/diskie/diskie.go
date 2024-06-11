package main

import (
	"bytes"
	"diskie"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/dustin/go-humanize"
	"github.com/urfave/cli"
)

var defaultFormat = `
{{
	printf "%-10s   %7s   %-15s   %-15s"
	( .Device | derefStr )
	( .PreferredSize | humanBytesIEC )
	( .IdType  | derefStr | default "-" | abbrev 15 )
	( .IdLabel | derefStr | default "-" | abbrev 15 )
}}
{{ with .RootDrive }}
   {{ .Model | derefStr | default "-" | compact }}
{{ else }}
   -
{{ end }}
`

func main() {
	app := &cli.App{
		Name:  "diskie",
		Usage: "Command line tool for UDisks2",
		Commands: []cli.Command{
			{
				Name:  "blockdevs",
				Usage: "Print block devices.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "format",
						Value: "json",
						Usage: `Output format. Can be "json", "default", or path to a file containing a golang template.`,
					},
					&cli.UintFlag{
						Name:  "min-importance",
						Value: 0,
						Usage: "Only include block devices more important than the given level. Possible values are 0 through 3.",
					},
					&cli.StringFlag{
						Name:  "json-type",
						Value: "array",
						Usage: `Type of the JSON output. Can be "array" or "object".`,
					},
				},
				Action: func(c *cli.Context) error {
					f := c.String("format")
					t := c.String("json-type")
					i := c.Uint("min-importance")
					if i > 3 {
						return fmt.Errorf("min-importance of %d is out of the possible range of 0 through 3", i)
					}
					return cmdBlockdevs(f, t, i)
				},
			},
			{
				Name:      "menu",
				Usage:     "Mount or unmount devices using a dmenu-compatible program.",
				UsageText: "menu [command options] cmd [arguments...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "format",
						Value: "default",
						Usage: `Output format. Can be "default" or path to a file containing a golang template.`,
					},
					&cli.UintFlag{
						Name:  "min-importance",
						Value: 0,
						Usage: "Only include block devices more important than the given level. Possible values are 0 through 3.",
					},
					&cli.StringFlag{
						Name:  "max-lines",
						Value: "0",
						Usage: "Limit the maximum value of the %l sequence. Zero means no limit.",
					},
				},
				Action: func(c *cli.Context) error {
					menuCmd := c.Args().First()
					menuArgs := c.Args().Tail()
					f := c.String("format")
					i := c.Uint("min-importance")
					l := c.Uint("max-lines")
					if i > 3 {
						return fmt.Errorf("min-importance of %d is out of the possible range of 0 through 3", i)
					}
					if menuCmd == "" && len(menuArgs) == 0 {
						return fmt.Errorf("please provide a dmenu-compatible program as the arguments to this command (eg. `diskie menu dmenu -p Diskie`)")
					}
					return cmdMenu(f, i, l, menuCmd, menuArgs)
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

func cmdBlockdevs(format string, jsonType string, importance uint) error {
	blocks, blockmap, err := blocks(importance)
	if err != nil {
		return err
	}

	if format == "json" {
		var v any

		if jsonType == "array" {
			v = blocks
		} else if jsonType == "object" {
			v = blockmap
		} else {
			return fmt.Errorf("unknown jsonType: %s", jsonType)
		}

		pretty, err := prettyJson(v)
		if err != nil {
			return err
		}

		fmt.Println(string(pretty))
		return nil
	} else if format == "default" {
		format = defaultFormat
	} else {
		f, err := os.ReadFile(format)
		if err != nil {
			return fmt.Errorf("could not read the format file: %w", err)
		}
		format = string(f)
	}

	formattedSlice, _, err := formatBlocks(blocks, format, false)
	if err != nil {
		return err
	}

	fmt.Println(strings.Join(formattedSlice, "\n"))
	return nil
}

func cmdMenu(format string, importance uint, maxlines uint, menuCmd string, menuArgs []string) error {
	blocks, _, err := blocks(importance)
	if err != nil {
		return err
	}

	if format == "default" {
		format = defaultFormat
	} else {
		f, err := os.ReadFile(format)
		if err != nil {
			return fmt.Errorf("could not read the format file: %w", err)
		}
		format = string(f)
	}

	formattedSlice, formattedMap, err := formatBlocks(blocks, format, true)
	if err != nil {
		return err
	}

	lines := uint(len(formattedSlice))
	if maxlines > 0 {
		lines = max(lines, maxlines)
	}
	linesStr := fmt.Sprint(lines)
	for i := 0; i < len(menuArgs); i++ {
		menuArgs[i] = strings.ReplaceAll(menuArgs[i], "%l", linesStr)
	}

	cmd := exec.Command(menuCmd, menuArgs...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("could not connect to the menu's stdin: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("could not connect to the menu's stdout: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("could not run the menu: %w", err)
	}

	_, err = stdin.Write([]byte(strings.Join(formattedSlice, "\n")))
	if err != nil {
		return fmt.Errorf("could not write to the menu's stdin: %w", err)
	}
	stdin.Close()

	var output bytes.Buffer

	go func() {
		defer stdout.Close()
		io.Copy(&output, stdout)
		if err != nil {
			panic(fmt.Errorf("could not read from the menu's stdout: %w", err))
		}
	}()

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("the menu command failed: %w", err)
	}

	selected := formattedMap[strings.TrimSuffix(output.String(), "\n")]
	pretty, err := prettyJson(selected)
	if err != nil {
		return err
	}

	fmt.Println(string(pretty))
	return nil
}

func formatBlocks(
	blocks []*diskie.BlockDevice, format string, catchDuplicate bool) (
	[]string, map[string]*diskie.BlockDevice, error) {

	funcs := template.FuncMap{
		"derefStr": deref[string],
		"derefU32": deref[uint32],
		"derefU64": deref[uint64],
		"compact": func(v string) string {
			return regexp.MustCompile(`\s+`).ReplaceAllString(v, " ")
		},
		"humanBytes": func(v uint64) string {
			return humanize.Bytes(v)
		},
		"humanBytesIEC": func(v uint64) string {
			return humanize.IBytes(v)
		},
	}

	tmpl, err := template.New("format").
		Funcs(sprig.FuncMap()).Funcs(funcs).Parse(format)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse the template: %w", err)
	}

	formattedSlice := make([]string, 0, len(blocks))
	formattedMap := make(map[string]*diskie.BlockDevice, len(blocks))

	for _, b := range blocks {
		var output bytes.Buffer
		err := tmpl.Execute(&output, b)
		if err != nil {
			return nil, nil, fmt.Errorf("could not execute the template: %w", err)
		}

		f := strings.ReplaceAll(output.String(), "\n", "")

		if catchDuplicate {
			_, has := formattedMap[f]
			if has {
				return nil, nil, fmt.Errorf("the output format leads to duplicates in the list of disks")
			}
		}

		formattedSlice = append(formattedSlice, f)
		formattedMap[f] = b
	}

	return formattedSlice, formattedMap, nil
}

func deref[T any](v *T) T {
	if v != nil {
		return *v
	} else {
		var v T
		return v
	}
}

func blocks(importance uint) (
	[]*diskie.BlockDevice, map[string]*diskie.BlockDevice, error) {
	dsk, err := diskie.Connect()
	if err != nil {
		return nil, nil, fmt.Errorf("could not create diskie client: %w", err)
	}

	blockmap, err := dsk.BlockDevices()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get block devices: %w", err)
	}

	blocks := blockmap.Sort()
	blocks, err = blockmap.Filter(blocks, importance)
	if err != nil {
		return nil, nil, err
	}

	return blocks, blockmap.BlockMap, nil
}

func prettyJson(obj interface{}) ([]byte, error) {
	output, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		return nil, fmt.Errorf("could not marshal object into json: %w", err)
	}
	return output, nil
}
