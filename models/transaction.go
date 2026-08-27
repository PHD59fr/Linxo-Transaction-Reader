package models

// Transaction represents a single bank transaction exported from Linxo.
type Transaction struct {
	From     string `json:"from"`
	Category string `json:"category,omitempty"`
	Amount   string `json:"amount"`
	Date     string `json:"date"`
	Note     string `json:"note,omitempty"`
}
