package linxo

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf16"

	"linxo-reader/models"
)

const (
	csvEndpoint = "https://wwws.linxo.com/secured/transactionsReport"
	refererURL  = "https://wwws.linxo.com/secured/history.page"
)

// DownloadCSV fetches the transaction CSV from Linxo using session cookies.
func DownloadCSV(cookies, viewID, userAgent string) ([]byte, error) {
	data := url.Values{
		"fileName":        {"opérations"},
		"extension":       {".csv"},
		"selectedViewIds": {viewID},
	}

	req, err := http.NewRequest("POST", csvEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookies)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", refererURL)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %.200s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ParseTransactions decodes a UTF-16LE CSV blob into Transaction records.
func ParseTransactions(csvData []byte) ([]models.Transaction, error) {
	text := decodeUTF16LE(csvData)

	records, err := parseTabSeparated(text)
	if err != nil {
		return nil, fmt.Errorf("parse CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, errors.New("CSV is empty")
	}

	header := records[0]
	colIdx := buildColumnIndex(header)

	dateCol := findColumn(colIdx, "date")
	labelCol := findColumn(colIdx, "libellé", "libelle")
	catCol := findColumn(colIdx, "catégor", "categor")
	amountCol := findColumn(colIdx, "montant")
	noteCol := findColumn(colIdx, "notes", "note")

	var rows []models.Transaction
	for _, rec := range records[1:] {
		if len(rec) < 3 {
			continue
		}
		label := getColValue(rec, labelCol)
		if label == "" {
			continue
		}
		rows = append(rows, models.Transaction{
			From:     label,
			Category: getColValue(rec, catCol),
			Amount:   getColValue(rec, amountCol),
			Date:     getColValue(rec, dateCol),
			Note:     getColValue(rec, noteCol),
		})
	}

	if len(rows) == 0 {
		return nil, errors.New("no transactions found")
	}
	return rows, nil
}

func decodeUTF16LE(data []byte) string {
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		data = data[2:]
	}
	u16 := make([]uint16, len(data)/2)
	for i := range u16 {
		u16[i] = uint16(data[i*2]) | uint16(data[i*2+1])<<8
	}
	return string(utf16.Decode(u16))
}

func parseTabSeparated(text string) ([][]string, error) {
	r := csv.NewReader(strings.NewReader(text))
	r.Comma = '\t'
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	return r.ReadAll()
}

func buildColumnIndex(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, h := range header {
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return idx
}

func findColumn(colMap map[string]int, names ...string) int {
	for _, name := range names {
		for col, idx := range colMap {
			if strings.Contains(col, name) {
				return idx
			}
		}
	}
	return -1
}

func getColValue(rec []string, idx int) string {
	if idx < 0 || idx >= len(rec) {
		return ""
	}
	return strings.TrimSpace(rec[idx])
}
