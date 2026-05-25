package dify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	host             string
	defaultAPISecret string
	httpClient       *http.Client
}

func NewClientWithConfig(c *ClientConfig) *Client {
	httpClient := &http.Client{}

	if c.Timeout != 0 {
		httpClient.Timeout = c.Timeout
	}
	if c.Transport != nil {
		httpClient.Transport = c.Transport
	}

	return &Client{
		host:             c.Host,
		defaultAPISecret: c.DefaultAPISecret,
		httpClient:       httpClient,
	}
}

func NewClient(host, defaultAPISecret string) *Client {
	return NewClientWithConfig(&ClientConfig{
		Host:             host,
		DefaultAPISecret: defaultAPISecret,
	})
}

func (c *Client) sendRequest(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *Client) sendJSONRequest(req *http.Request, res interface{}) error {
	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Status  int    `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return fmt.Errorf("HTTP %d: failed to decode error response", resp.StatusCode)
		}
		return fmt.Errorf("HTTP response error: [%v]%v", errBody.Code, errBody.Message)
	}

	if res != nil {
		if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) sendDeleteRequest(req *http.Request) error {
	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		var errBody struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Status  int    `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return fmt.Errorf("HTTP %d: failed to decode error response", resp.StatusCode)
		}
		return fmt.Errorf("HTTP response error: [%v]%v", errBody.Code, errBody.Message)
	}
	return nil
}

func (c *Client) getHost() string {
	return strings.TrimSuffix(c.host, "/")
}

func (c *Client) getAPISecret() string {
	return c.defaultAPISecret
}

func (c *Client) API() *API {
	return &API{
		c: c,
	}
}
