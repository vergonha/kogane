package turnstile

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const verifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type Response struct {
	Success bool     `json:"success"`
	Errors  []string `json:"error-codes"`
}

type Client struct {
	SecretKey  string
	HTTPClient *http.Client
}

func New(secretKey string) *Client {
	return &Client{
		SecretKey: secretKey,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) Verify(r *http.Request) bool {
	if c.SecretKey == "" {
		log.Fatal("CLOUDFLARE_TURNSTILE_SECRET_KEY not configured")
		return false
	}

	token := r.FormValue("cf-turnstile-response")
	if token == "" {
		return false
	}

	form := url.Values{}
	form.Set("secret", c.SecretKey)
	form.Set("response", token)

	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		form.Set("remoteip", ip)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		verifyURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return false
	}

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Turnstile: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	return result.Success
}
