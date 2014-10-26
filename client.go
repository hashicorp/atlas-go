package harmony

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

// Client represents a single connection to a Harmony API endpoint.
type Client struct {
	// URL is the full endpoint address to the Harmony server including the
	// protocol, port, and path.
	URL    *url.URL
	rawURL string

	// token is the Harmony authentication token
	token string

	// httpClient is the underlying http client with which to make requests.
	httpClient *http.Client
}

//
func NewClient(url string) (*Client, error) {
	client := &Client{rawURL: url}
	if err := client.init(); err != nil {
		return nil, err
	}
	return client, nil
}

//
func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

//
func (c *Client) SetToken(token string) {
	c.token = token
}

//
func (c *Client) init() error {
	if len(c.rawURL) == 0 {
		return fmt.Errorf("client: misisng url")
	}

	u, err := url.Parse(c.rawURL)
	if err != nil {
		return err
	}
	c.URL = u

	c.httpClient = http.DefaultClient

	return nil
}

//
// If a non-200 status code is returned, it is assumed to be an error and an
// error is returned. The http.Response object is also returned, so furher
// inspection may be done to determine the response and appropriate action.
func (c *Client) request(verb, thepath string) (*http.Response, error) {
	u := *c.URL
	u.Path = path.Join(c.URL.Path, thepath)

	request, err := http.NewRequest(verb, u.String(), nil)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		var buff bytes.Buffer
		io.Copy(&buff, response.Body)
		return response, fmt.Errorf("client: unexpected response code %d:\n%s",
			response.StatusCode, buff.Bytes())
	}

	return response, nil
}
