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
	URL *url.URL

	// Token is the Harmony authentication token
	Token string

	// HTTPClient is the underlying http client with which to make requests.
	HTTPClient *http.Client
}

// NewClient creates a new Harmony Client from the given URL (as a string). If
// the URL cannot be parsed, an error is returned. The HTTPClient is set to
// http.DefaultClient, but this can be changed programatically by setting
// client.HTTPClient. The user can also programtically set the URL as a
// *url.URL.
func NewClient(urlString string) (*Client, error) {
	if len(urlString) == 0 {
		return nil, fmt.Errorf("client: missing url")
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	client := &Client{URL: parsedURL}
	if err := client.init(); err != nil {
		return nil, err
	}

	return client, nil
}

// Login accepts a username and password as string arguments. Both username and
// password must be non-nil, non-empty values. Harmony does not permit
// passwordless authentication.
//
// If authentication is unsuccessful, an error is returned with the body of the
// error containing the server's response.
//
// If authentication is successful, this method sets the Token value on the
// Client and returns the Token as a string.
func (c *Client) Login(username, password string) (string, error) {
	if len(username) == 0 {
		return "", fmt.Errorf("client: missing username")
	}

	if len(password) == 0 {
		return "", fmt.Errorf("client: missing password")
	}

	// Make a request

	// If there's an error, return it

	// Set the token

	// Return the token
	return "", nil
}

// init() sets defaults on the client.
func (c *Client) init() error {
	c.HTTPClient = http.DefaultClient
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

	response, err := c.HTTPClient.Do(request)
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
