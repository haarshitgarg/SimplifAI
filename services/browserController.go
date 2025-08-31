package services

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// Get the HTML of a given URL
func (p *webParser) getHTML(rawURL string) (string, error) {
	fmt.Printf("getHTML function called to get html for %s\n", rawURL)

	var html_str string
	err := chromedp.Run(
		p.ctx,
		chromedp.Navigate(rawURL),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &html_str),
	)
	if err != nil {
		return "", fmt.Errorf("Could not find the html body for url: %s\nError: %s\n", rawURL, err)
	}

	return html_str, nil
}

// Open a new tab with a given html
func (p *webParser) openTabWithHtml(rawHTML string) error {
	fmt.Println("Called openTabWithHtml to open html file in new tab")
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("start-maximized", true),
	)
	allocCtx, _ := chromedp.NewExecAllocator(p.ctx, opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	encoded := base64.StdEncoding.EncodeToString([]byte(rawHTML))
	dataURL := "data:text/html;base64,"+encoded
	chromedp.Run(
		ctx,
		chromedp.Navigate(dataURL),
		chromedp.Sleep(50*time.Second),
	)

	return nil
}

