package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/mitchellh/go-homedir"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Flag struct {
	Key      string
	DefValue interface{}
}

func InitConfig(v *viper.Viper, file string, defDir string, name string) func() {
	return func() {
		if file != "" {
			v.SetConfigFile(file)
		} else {
			home, err := homedir.Dir()
			if err != nil {
				panic(err)
			}
			v.AddConfigPath(path.Join("./", defDir)) // local config takes priority
			v.AddConfigPath(path.Join(home, defDir))
			v.SetConfigName(name)
		}

		v.SetEnvPrefix("TXTL")
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()
		_ = v.ReadInConfig()
	}
}

func BindFlags(v *viper.Viper, root *cobra.Command, flags map[string]Flag) error {
	for n, f := range flags {
		if err := v.BindPFlag(f.Key, root.PersistentFlags().Lookup(n)); err != nil {
			return err
		}
		v.SetDefault(f.Key, f.DefValue)
	}
	return nil
}

func ExpandConfigVars(v *viper.Viper, flags map[string]Flag) {
	for _, f := range flags {
		if f.Key != "" {
			if str, ok := v.Get(f.Key).(string); ok {
				v.Set(f.Key, os.ExpandEnv(str))
			}
		}
	}
}

func AddrFromStr(str string) ma.Multiaddr {
	addr, err := ma.NewMultiaddr(str)
	if err != nil {
		Fatal(err)
	}
	return addr
}

func Message(format string, args ...interface{}) {
	fmt.Println(aurora.Sprintf(aurora.BrightBlack("> "+format), args...))
}

func Success(format string, args ...interface{}) {
	fmt.Println(aurora.Sprintf(aurora.Cyan("> Success! %s"),
		aurora.Sprintf(aurora.BrightBlack(format), args...)))
}

func Fatal(err error, args ...interface{}) {
	words := strings.SplitN(err.Error(), " ", 2)
	words[0] = strings.Title(words[0])
	msg := strings.Join(words, " ")
	fmt.Println(aurora.Sprintf(aurora.Red("> Error! %s"),
		aurora.Sprintf(aurora.BrightBlack(msg), args...)))
	os.Exit(1)
}

func RenderTable(header []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	headersColors := make([]tablewriter.Colors, len(data[0]))
	for i := range headersColors {
		headersColors[i] = tablewriter.Colors{tablewriter.FgHiBlackColor}
	}
	table.SetHeaderColor(headersColors...)
	table.AppendBulk(data)
	table.Render()
	fmt.Println()
}

func checkErr(e error) {
	if e != nil {
		Fatal(e)
	}
}
