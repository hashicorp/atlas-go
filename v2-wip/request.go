package atlas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	pathlib "path"
)

// RequestOptions is the list of options to pass to the request.
type RequestOptions struct {
	// Params is a map of key-value pairs that will be added to the Request URL.
	Params map[string]string

	// Headers is a map of key-value pairs that will be added to the Request.
	Headers map[string]string

	// Body is an io.Reader object that will be streamed or uploaded with the
	// Request. BodyLength is the final size of the Body.
	Body       io.Reader
	BodyLength int64
}

// Request performs an HTTP request with the given options.
//
// The request is modified to have the proper auth token, content type, and
// so on. The response is returned as-is and should be handled as a normal
// HTTP response.
//
// The response error codes will be checked according to the JSON API
// standard. That means that if an error is returned, it may be an *ErrorDoc
// with further details. Type casting can be used to test this.
func (c *Client) Request(verb, path string, ro *RequestOptions) (*http.Response, error) {
	req, err := c.RawRequest(verb, path, ro)
	if err != nil {
		return nil, err
	}

	rawResp, err := checkResp(c.HTTPClient.Do(req))
	if err != nil {
		return nil, err
	}

	return rawResp, nil
}

// RawRequest takes RequestOptions and returns a raw http.Request struct.
// This can then be used to manually make requests.
func (c *Client) RawRequest(verb, path string, ro *RequestOptions) (*http.Request, error) {
	if ro == nil {
		ro = new(RequestOptions)
	}

	log.Printf("[INFO] atlas/v2: building request: %q %q", verb, path)

	// Create a new URL with the appended path
	u := *c.URL
	u.Path = pathlib.Join(c.URL.Path, path)

	// Add the token and other params
	if c.Token != "" {
		log.Printf("[DEBUG] atlas/v2: appending token (%s)", maskString(c.Token))
		if ro.Headers == nil {
			ro.Headers = make(map[string]string)
		}

		ro.Headers[atlasTokenHeader] = c.Token
	}

	// Add the token and other params
	var params = make(url.Values)
	for k, v := range ro.Params {
		params.Add(k, v)
	}
	u.RawQuery = params.Encode()

	// Create the request object
	request, err := http.NewRequest(verb, u.String(), ro.Body)
	if err != nil {
		return nil, err
	}

	// set our default headers first
	for k, v := range c.DefaultHeader {
		request.Header[k] = v
	}

	// Add any request headers (auth will be here if set)
	for k, v := range ro.Headers {
		request.Header.Add(k, v)
	}

	// Add content-length if we have it
	if ro.BodyLength > 0 {
		request.ContentLength = ro.BodyLength
	}

	log.Printf("[DEBUG] atlas/v2: raw request: %#v", request)
	return request, nil
}

// checkResp wraps http.Client.Do() and verifies that the request was
// successful. A non-200 request returns an error formatted to included any
// validation problems or otherwise.
func checkResp(resp *http.Response, err error) (*http.Response, error) {
	// If the err is already there, there was an error higher up the chain, so
	// just return that
	if err != nil {
		return resp, err
	}

	log.Printf("[INFO] atlas/v2: response: %d (%s)", resp.StatusCode, resp.Status)

	// Load the body into memory. The Atlas API never responds with a huge
	// response so this helps save some burden on clients (they don't have
	// to close responses).
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		log.Printf("[ERR] atlas/v2: error copying response body: %s", err)
		return nil, fmt.Errorf("error copying response body: %s", err)
	} else {
		log.Printf("[DEBUG] atlas/v2: response: %s", buf.String())

		// We are going to reset the response body, so we need to close the
		// old one or else it will leak.
		resp.Body.Close()
		resp.Body = &bytesReadCloser{Buffer: &buf}
	}

	// Check for any errors
	switch resp.StatusCode {
	case 200, 201, 202, 203, 204:
		return resp, nil
	case 400:
		return nil, parseErr(resp)
	case 401:
		return nil, ErrAuth
	case 404:
		return nil, ErrNotFound
	case 422:
		return nil, parseErr(resp)
	default:
		return nil, fmt.Errorf("client: unexected status code %d", resp.StatusCode)
	}
}

// parseErr is used to take an error JSON response and return a single string
// for use in error messages.
func parseErr(r *http.Response) error {
	var doc ErrorDoc
	err := json.NewDecoder(r.Body).Decode(&doc)
	r.Body.Close()
	if err != nil {
		return fmt.Errorf("error decoding JSON body: %s", err)
	}

	return &doc
}

// bytesReadCloser is a simple wrapper around a bytes buffer that implements
// Close as a noop.
type bytesReadCloser struct {
	*bytes.Buffer
}

func (nrc *bytesReadCloser) Close() error {
	// we don't actually have to do anything here, since the buffer is just some
	// data in memory  and the error is initialized to no-error
	return nil
}
