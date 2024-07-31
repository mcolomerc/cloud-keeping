package table

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

func NewTable(header []string, rows [][]interface{}) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	headerRow := table.Row{}
	for _, h := range header {
		headerRow = append(headerRow, h)
	}
	t.AppendHeader(headerRow)

	for _, row := range rows {
		t.AppendRow(row)
	}
	t.AppendSeparator()
	t.AppendFooter(table.Row{"Total", len(rows)})

	t.SetStyle(table.StyleRounded)
	t.Render()
}
