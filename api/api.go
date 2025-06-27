package api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

// Client holds the base configuration for making HTTP requests.
type Client struct {
	BaseURL  string
	Username string
	Password string
	APIToken string
}

// NewClient creates a new API client.
func NewClient(baseURL, username, password, token string) *Client {
	return &Client{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		APIToken: token,
	}
}

// MakeRequest makes an HTTP request and returns the response body.
// We are just returing the bytes in this method. The actual coversion to
// a struct can be done in the caller.
func (c *Client) MakeRequest(method, url string, payload []byte) ([]byte, error) {
	client := &http.Client{}

	// Construct the full URL
	fullURL := c.BaseURL + url

	// Set up the request
	req, err := http.NewRequest(method, fullURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	// Encode credentials
	// For JIRA API tokens, use Basic auth with username:token
	var auth string
	if c.APIToken != "" {
		auth = base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.APIToken))
	} else {
		auth = base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))
	}
	req.Header.Add("Authorization", "Basic "+auth)

	if method == "POST" || method == "PUT" {
		req.Header.Add("Content-Type", "application/json")
	}
	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Log response for debugging
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
