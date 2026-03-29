package linxo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"linxo-reader/internal/browser"
	"linxo-reader/internal/config"
	"linxo-reader/models"
)

const linxoScopeURL = "https://wwws.linxo.com"

// FetchTransactions orchestrates the full flow: browser login, cookie
// extraction, CSV download, and parsing.
func FetchTransactions(ctx context.Context, cfg *config.Config) ([]models.Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, cfg.BrowserTimeout)
	defer cancel()

	sess, err := browser.New(ctx, cfg.Headless, cfg.ChromeBin)
	if err != nil {
		return nil, fmt.Errorf("browser: %w", err)
	}
	defer sess.Close()

	if err := sess.Navigate(loginURL); err != nil {
		return nil, err
	}
	time.Sleep(2 * time.Second)

	if err := Login(sess.Page, cfg.LinxoEmail, cfg.LinxoPass); err != nil {
		return nil, err
	}

	cookieParts, viewID, err := sess.Cookies(linxoScopeURL)
	if err != nil {
		return nil, err
	}
	if viewID == "" {
		return nil, errors.New("LinxoPViewSelection cookie not found")
	}

	userAgent, err := sess.UserAgent()
	if err != nil {
		return nil, err
	}

	csvData, err := DownloadCSV(cookieParts[0], viewID, userAgent)
	if err != nil {
		return nil, fmt.Errorf("download CSV: %w", err)
	}
	log.Printf("Downloaded CSV: %d bytes", len(csvData))

	return ParseTransactions(csvData)
}
