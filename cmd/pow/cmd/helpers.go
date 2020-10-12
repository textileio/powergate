package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"github.com/textileio/powergate/api/client"
)

// Message prints a message to stdout.
func Message(format string, args ...interface{}) {
	fmt.Println(aurora.Sprintf(aurora.BrightBlack("> "+format), args...))
}

// Success prints a success message to stdout.
func Success(format string, args ...interface{}) {
	fmt.Println(aurora.Sprintf(aurora.Cyan("> Success! %s"),
		aurora.Sprintf(aurora.BrightBlack(format), args...)))
}

// Fatal prints a fatal error to stdout, and exits immediately with
// error code 1.
func Fatal(err error, args ...interface{}) {
	words := strings.SplitN(err.Error(), " ", 2)
	words[0] = strings.Title(words[0])
	msg := strings.Join(words, " ")
	fmt.Println(aurora.Sprintf(aurora.Red("> Error! %s"),
		aurora.Sprintf(aurora.BrightBlack(msg), args...)))
	os.Exit(1)
}

// RenderTable renders a table with header columns and data rows to writer.
func RenderTable(writer io.Writer, header []string, data [][]string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	headersColors := make([]tablewriter.Colors, len(header))
	for i := range headersColors {
		headersColors[i] = tablewriter.Colors{tablewriter.FgHiBlackColor}
	}
	table.SetHeaderColor(headersColors...)
	table.AppendBulk(data)
	table.Render()
}

func checkErr(e error) {
	if e != nil {
		Fatal(e)
	}
}

func mustAuthCtx(ctx context.Context) context.Context {
	token := viper.GetString("token")
	if token == "" {
		Fatal(errors.New("must provide -t token"))
	}
	return context.WithValue(ctx, client.AuthKey, token)
}

func authCtx(ctx context.Context) context.Context {
	token := viper.GetString("token")
	return context.WithValue(ctx, client.AuthKey, token)
}
