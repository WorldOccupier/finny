package importpreview

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

func readRows(data []byte, filename string) ([]inputRow, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case FILE_EXTENSION_CSV:
		reader := csv.NewReader(bytes.NewReader(data))
		rows, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("read CSV: %w", err)
		}
		if len(rows) < 2 {
			return nil, fmt.Errorf("statement must contain a header and at least one row")
		}
		return nonBlankRows(rows[1:], 2), nil
	case FILE_EXTENSION_XLSX:
		file, err := excelize.OpenReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("read XLSX: %w", err)
		}
		sheets := file.GetSheetMap()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("XLSX contains no sheets")
		}
		rows, err := file.Rows(sheets[1])
		if err != nil {
			return nil, fmt.Errorf("read XLSX rows: %w", err)
		}
		var result []inputRow
		sourceRow := 0
		for rows.Next() {
			sourceRow++
			cells, err := rows.Columns()
			if err != nil {
				return nil, fmt.Errorf("read XLSX cells: %w", err)
			}
			result = append(result, inputRow{sourceRow: sourceRow, cells: cells})
		}
		if err := rows.Error(); err != nil {
			return nil, fmt.Errorf("read XLSX rows: %w", err)
		}
		if len(result) < 1 {
			return nil, fmt.Errorf("statement must contain a header and at least one row")
		}
		return nonBlankInputRows(result[1:]), nil
	default:
		return nil, fmt.Errorf("unsupported statement format")
	}
}

func nonBlankRows(rows [][]string, firstSourceRow int) []inputRow {
	result := make([]inputRow, 0, len(rows))
	for index, cells := range rows {
		if isBlankRow(cells) {
			continue
		}
		result = append(result, inputRow{sourceRow: firstSourceRow + index, cells: cells})
	}
	return result
}

func nonBlankInputRows(rows []inputRow) []inputRow {
	result := rows[:0]
	for _, row := range rows {
		if !isBlankRow(row.cells) {
			result = append(result, row)
		}
	}
	return result
}

func isBlankRow(cells []string) bool {
	for _, cell := range cells {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
