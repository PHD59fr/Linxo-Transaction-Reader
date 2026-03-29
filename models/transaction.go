package models

// Transaction represents a single bank transaction exported from Linxo.
type Transaction struct {
	From     string `json:"from" csv:"Libellé"`
	Category string `json:"category,omitempty" csv:"Catégorie"`
	Amount   string `json:"amount" csv:"Montant"`
	Date     string `json:"date" csv:"Date"`
	Note     string `json:"note,omitempty" csv:"Notes"`
}

// CSVRecord holds raw column indices mapped to known CSV headers.
type CSVRecord struct {
	DateCol     int
	LabelCol    int
	CategoryCol int
	AmountCol   int
	NoteCol     int
}
