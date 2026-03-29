package browser

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Session holds a running browser instance and its active page.
type Session struct {
	Browser *rod.Browser
	Page    *rod.Page
}

// New launches a Chromium instance and returns a connected Session.
func New(ctx context.Context, headless bool, chromeBin string) (*Session, error) {
	l := launcher.New().
		Headless(headless).
		NoSandbox(true).
		Set("disable-gpu", "true")

	if chromeBin != "" {
		l = l.Bin(chromeBin)
	}

	wsURL := l.MustLaunch()

	browser := rod.New().ControlURL(wsURL).Context(ctx)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("browser connect: %w", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		_ = browser.Close()
		return nil, fmt.Errorf("page creation: %w", err)
	}

	page.MustSetViewport(1400, 1200, 1, false)

	return &Session{Browser: browser, Page: page}, nil
}

// Navigate loads a URL and waits for the page to finish loading.
func (s *Session) Navigate(url string) error {
	if err := s.Page.Navigate(url); err != nil {
		return fmt.Errorf("navigate: %w", err)
	}
	s.Page.MustWaitLoad()
	return nil
}

// Cookies returns all cookies for the given URL scope.
func (s *Session) Cookies(scopeURL string) ([]string, string, error) {
	cookies, err := s.Page.Cookies([]string{scopeURL})
	if err != nil {
		return nil, "", fmt.Errorf("get cookies: %w", err)
	}

	cookieStr := ""
	viewID := ""
	for _, c := range cookies {
		cookieStr += c.Name + "=" + c.Value + "; "
		if c.Name == "LinxoPViewSelection" {
			viewID = c.Value
		}
	}
	log.Printf("Got %d cookies", len(cookies))

	return []string{cookieStr}, viewID, nil
}

// UserAgent returns the browser's current user-agent string.
func (s *Session) UserAgent() (string, error) {
	res, err := s.Page.Eval(`() => navigator.userAgent`)
	if err != nil {
		return "", fmt.Errorf("eval user-agent: %w", err)
	}
	return res.Value.Str(), nil
}

// Close tears down the page and browser.
func (s *Session) Close() {
	if s.Page != nil {
		_ = s.Page.Close()
	}
	if s.Browser != nil {
		_ = s.Browser.Close()
	}
}

func init() {
	if os.Getenv("ROD_BROWSER_PATH") == "" && os.Getenv("CHROME_BIN") != "" {
		_ = os.Setenv("ROD_BROWSER_PATH", os.Getenv("CHROME_BIN"))
	}
}
