package linxo

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

// helpers ------------------------------------------------------------------

// encodeUTF16LE encodes a Go string as UTF-16LE bytes with optional BOM.
func encodeUTF16LE(s string, withBOM bool) []byte {
	encoded := utf16.Encode([]rune(s))
	buf := make([]byte, len(encoded)*2)
	for i, v := range encoded {
		binary.LittleEndian.PutUint16(buf[i*2:], v)
	}
	if withBOM {
		return append([]byte{0xFF, 0xFE}, buf...)
	}
	return buf
}

// makeTabCSV builds a UTF-16LE tab-separated CSV with a BOM.
func makeTabCSV(rows ...string) []byte {
	text := ""
	for i, row := range rows {
		if i > 0 {
			text += "\n"
		}
		text += row
	}
	return encodeUTF16LE(text, true)
}

// decodeUTF16LE tests -------------------------------------------------------

func TestDecodeUTF16LE_WithBOM(t *testing.T) {
	input := encodeUTF16LE("hello", true)
	got := decodeUTF16LE(input)
	if got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
}

func TestDecodeUTF16LE_WithoutBOM(t *testing.T) {
	input := encodeUTF16LE("world", false)
	got := decodeUTF16LE(input)
	if got != "world" {
		t.Errorf("expected %q, got %q", "world", got)
	}
}

func TestDecodeUTF16LE_Empty(t *testing.T) {
	got := decodeUTF16LE(nil)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestDecodeUTF16LE_Unicode(t *testing.T) {
	want := "café ñ €"
	input := encodeUTF16LE(want, true)
	got := decodeUTF16LE(input)
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

// parseTabSeparated tests ---------------------------------------------------

func TestParseTabSeparated_Basic(t *testing.T) {
	text := "a\tb\tc\n1\t2\t3"
	got, err := parseTabSeparated(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(got))
	}
	if len(got[0]) != 3 || got[0][0] != "a" || got[0][2] != "c" {
		t.Errorf("unexpected header row: %v", got[0])
	}
	if len(got[1]) != 3 || got[1][1] != "2" {
		t.Errorf("unexpected data row: %v", got[1])
	}
}

func TestParseTabSeparated_LazyQuotes(t *testing.T) {
	text := `a\t"b with "quotes""`
	_, err := parseTabSeparated(text)
	if err != nil {
		t.Fatalf("lazy quotes should not error: %v", err)
	}
}

// buildColumnIndex tests ----------------------------------------------------

func TestBuildColumnIndex(t *testing.T) {
	header := []string{"Date", "Libellé", "Montant", "Notes"}
	idx := buildColumnIndex(header)

	if idx["date"] != 0 {
		t.Errorf("expected date at 0, got %d", idx["date"])
	}
	if idx["libellé"] != 1 {
		t.Errorf("expected libellé at 1, got %d", idx["libellé"])
	}
	if idx["montant"] != 2 {
		t.Errorf("expected montant at 2, got %d", idx["montant"])
	}
	if idx["notes"] != 3 {
		t.Errorf("expected notes at 3, got %d", idx["notes"])
	}
}

func TestBuildColumnIndex_TrimsSpaces(t *testing.T) {
	header := []string{"  Date  ", " Libellé "}
	idx := buildColumnIndex(header)
	if idx["date"] != 0 {
		t.Errorf("expected date trimmed to 0, got %d", idx["date"])
	}
}

// findColumn tests ----------------------------------------------------------

func TestFindColumn_ExactMatch(t *testing.T) {
	colMap := map[string]int{"date": 0, "libellé": 1, "montant": 2}
	if got := findColumn(colMap, "date"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestFindColumn_PartialMatch(t *testing.T) {
	colMap := map[string]int{"catégorie de transaction": 3}
	if got := findColumn(colMap, "catégor"); got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
}

func TestFindColumn_NoMatch(t *testing.T) {
	colMap := map[string]int{"date": 0}
	if got := findColumn(colMap, "nonexistent"); got != -1 {
		t.Errorf("expected -1, got %d", got)
	}
}

func TestFindColumn_MultipleNames(t *testing.T) {
	colMap := map[string]int{"libelle": 1}
	if got := findColumn(colMap, "libellé", "libelle"); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}

// getColValue tests ---------------------------------------------------------

func TestGetColValue_Valid(t *testing.T) {
	rec := []string{"a", "b", "c"}
	if got := getColValue(rec, 1); got != "b" {
		t.Errorf("expected %q, got %q", "b", got)
	}
}

func TestGetColValue_TrimsWhitespace(t *testing.T) {
	rec := []string{"  padded  "}
	if got := getColValue(rec, 0); got != "padded" {
		t.Errorf("expected %q, got %q", "padded", got)
	}
}

func TestGetColValue_NegativeIndex(t *testing.T) {
	rec := []string{"a"}
	if got := getColValue(rec, -1); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestGetColValue_OutOfBounds(t *testing.T) {
	rec := []string{"a"}
	if got := getColValue(rec, 5); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// ParseTransactions tests ---------------------------------------------------

func TestParseTransactions_ValidCSV(t *testing.T) {
	csvData := makeTabCSV(
		"Date\tLibellé\tCatégorie\tMontant\tNotes",
		"25/03/2026\tCARTE MONOPRIX\tCourses\t-42.50\t",
		"24/03/2026\tVIREMENT SALAIRE\tRevenus\t2500.00\tSalaire mars",
	)

	rows, err := ParseTransactions(csvData)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(rows))
	}

	if rows[0].From != "CARTE MONOPRIX" {
		t.Errorf("expected From=%q, got %q", "CARTE MONOPRIX", rows[0].From)
	}
	if rows[0].Amount != "-42.50" {
		t.Errorf("expected Amount=%q, got %q", "-42.50", rows[0].Amount)
	}
	if rows[0].Date != "25/03/2026" {
		t.Errorf("expected Date=%q, got %q", "25/03/2026", rows[0].Date)
	}
	if rows[0].Category != "Courses" {
		t.Errorf("expected Category=%q, got %q", "Courses", rows[0].Category)
	}

	if rows[1].Note != "Salaire mars" {
		t.Errorf("expected Note=%q, got %q", "Salaire mars", rows[1].Note)
	}
}

func TestParseTransactions_EmptyCSV(t *testing.T) {
	csvData := makeTabCSV("Date\tLibellé\tMontant")
	_, err := ParseTransactions(csvData)
	if err == nil {
		t.Fatal("expected error for empty CSV")
	}
}

func TestParseTransactions_InvalidUTF8(t *testing.T) {
	_, err := ParseTransactions([]byte{0x80, 0x81})
	if err == nil {
		t.Fatal("expected error for invalid data")
	}
}

func TestParseTransactions_SkipsShortRows(t *testing.T) {
	csvData := makeTabCSV(
		"Date\tLibellé\tMontant",
		"25/03/2026\tVALID\t10.00",
		"short",
		"24/03/2026\tALSO VALID\t20.00",
	)

	rows, err := ParseTransactions(csvData)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 transactions (short row skipped), got %d", len(rows))
	}
}

func TestParseTransactions_SkipsEmptyLabel(t *testing.T) {
	csvData := makeTabCSV(
		"Date\tLibellé\tMontant",
		"25/03/2026\t\t10.00",
		"24/03/2026\tVALID\t20.00",
	)

	rows, err := ParseTransactions(csvData)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 transaction (empty label skipped), got %d", len(rows))
	}
}

func TestParseTransactions_AccentedHeaders(t *testing.T) {
	csvData := makeTabCSV(
		"Date\tLibellé\tCatégorie\tMontant\tNotes",
		"01/01/2026\tTEST\tAlimentation\t-9.99\tune note",
	)

	rows, err := ParseTransactions(csvData)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1, got %d", len(rows))
	}
	if rows[0].Category != "Alimentation" {
		t.Errorf("expected %q, got %q", "Alimentation", rows[0].Category)
	}
}
