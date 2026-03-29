package linxo

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

const loginURL = "https://wwws.linxo.com/auth.page#Login"

// Login authenticates on the Linxo login page.
func Login(page *rod.Page, email, password string) error {
	log.Println("Starting login")

	emailEl, err := findFirst(page, 5*time.Second,
		`input[name="username"]`,
		`input[type="email"]`,
		`input[type="text"]`,
	)
	if err != nil {
		return fmt.Errorf("email field not found: %w", err)
	}
	if err := emailEl.Input(email); err != nil {
		return fmt.Errorf("email input: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	passEl, err := findFirst(page, 8*time.Second,
		`input[name="password"]`,
		`input[type="password"]`,
	)
	if err != nil {
		return fmt.Errorf("password field not found: %w", err)
	}
	if err := passEl.Input(password); err != nil {
		return fmt.Errorf("password input: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	btn, err := findFirst(page, 3*time.Second, `button[type="submit"]`, `button`)
	if err == nil && btn != nil {
		_ = btn.Click(proto.InputMouseButtonLeft, 1)
	} else {
		_ = page.Keyboard.Press(input.Enter)
	}

	page.MustWaitLoad()
	time.Sleep(4 * time.Second)
	log.Println("Login completed")
	return nil
}

func findFirst(page *rod.Page, timeout time.Duration, selectors ...string) (*rod.Element, error) {
	for _, sel := range selectors {
		el, err := page.Timeout(timeout).Element(sel)
		if err == nil && el != nil {
			return el, nil
		}
	}
	return nil, errors.New("no selector matched")
}
