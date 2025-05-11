package logger

import (
	"fmt"
	"log/slog"
	"strings"
)

type Table struct {
	headers     []string
	rows        [][]string
	columnWidth []int
	logger      *slog.Logger
}

func NewTable(headers []string, logger *slog.Logger) *Table {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	return &Table{
		headers:     headers,
		columnWidth: widths,
		logger:      logger,
	}
}

func (t *Table) AddRow(cells ...string) {
	if len(cells) > len(t.headers) {
		cells = cells[:len(t.headers)]
	} else if len(cells) < len(t.headers) {
		padded := make([]string, len(t.headers))
		copy(padded, cells)
		cells = padded
	}

	for i, cell := range cells {
		if len(cell) > t.columnWidth[i] {
			t.columnWidth[i] = len(cell)
		}
	}

	t.rows = append(t.rows, cells)
}

func (t *Table) Print() {
	var sb strings.Builder

	format := "│"
	separator := "├"
	footer := "└"

	for _, width := range t.columnWidth {
		format += " %-" + fmt.Sprintf("%d", width) + "s │"
		separator += strings.Repeat("─", width+2) + "┼"
		footer += strings.Repeat("─", width+2) + "┴"
	}

	separator = separator[:len(separator)-1] + "┤"
	footer = footer[:len(footer)-1] + "┘"

	sb.WriteString("┌")
	for i, width := range t.columnWidth {
		sb.WriteString(strings.Repeat("─", width+2))
		if i < len(t.columnWidth)-1 {
			sb.WriteString("┬")
		}
	}
	sb.WriteString("┐\n")

	headerArgs := make([]interface{}, len(t.headers))
	for i, h := range t.headers {
		headerArgs[i] = h
	}
	sb.WriteString(fmt.Sprintf(format, headerArgs...))
	sb.WriteString("\n")

	sb.WriteString(separator)
	sb.WriteString("\n")

	for _, row := range t.rows {
		args := make([]interface{}, len(row))
		for i, cell := range row {
			args[i] = cell
		}
		sb.WriteString(fmt.Sprintf(format, args...))
		sb.WriteString("\n")
	}
	sb.WriteString(footer)
	fmt.Println(sb.String())
}
