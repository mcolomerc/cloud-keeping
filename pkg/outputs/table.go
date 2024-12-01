package outputs

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

/*
*
Each element in the rows slice is a string that will be added to the first column of the table.
The header contains the header of the row.
*/
func BuildList(rows []string, header string) {
	y := make([][]interface{}, len(rows))
	for i, value := range rows {
		row := make([]interface{}, 1)
		row[0] = value
		y[i] = row
	}
	NewTable([]string{header}, y)
}
