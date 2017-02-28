package atlas

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-rootcerts"
)

const (
	// atlasDefaultEndpoint is the default base URL for connecting to Atlas.
	atlasDefaultEndpoint = "https://atlas.hashicorp.com"

	// atlasEndpointEnvVar is the environment variable that overrrides the
	// default Atlas address.
	atlasEndpointEnvVar = "ATLAS_ADDRESS"

	// atlasCAFileEnvVar is the environment variable that causes the client to
	// load trusted certs from a file
	atlasCAFileEnvVar = "ATLAS_CAFILE"

	// atlasCAPathEnvVar is the environment variable that causes the client to
	// load trusted certs from a directory
	atlasCAPathEnvVar = "ATLAS_CAPATH"

	// atlasTLSNoVerifyEnvVar disables TLS verification, similar to curl -k
	// This defaults to false (verify) and will change to true (skip
	// verification) with any non-empty value
	atlasTLSNoVerifyEnvVar = "ATLAS_TLS_NOVERIFY"

	// atlasTokenHeader is the header key used for authenticating with Atlas
	atlasTokenHeader = "X-Atlas-Token"
)

var projectURL = "https://github.com/hashicorp/atlas-go"
var userAgent = fmt.Sprintf("AtlasGo/2.0 (+%s; %s)",
	projectURL, runtime.Version())

// Client represents a single connection to a Atlas API endpoint.
type Client struct {
	// URL is the full endpoint address to the Atlas server including the
	// protocol, port, and path.
	URL *url.URL

	// Token is the Atlas authentication token
	Token string

	// HTTPClient is the underlying http client with which to make requests.
	HTTPClient *http.Client

	// DefaultHeaders is a set of headers that will be added to every request.
	// This minimally includes the atlas user-agent string.
	DefaultHeader http.Header
}

// DefaultClient returns a client that connects to the Atlas API.
func DefaultClient() *Client {
	atlasEndpoint := os.Getenv(atlasEndpointEnvVar)
	if atlasEndpoint == "" {
		atlasEndpoint = atlasDefaultEndpoint
	}

	client, err := NewClient(atlasEndpoint)
	if err != nil {
		panic(err)
	}

	return client
}

// NewClient creates a new Atlas Client from the given URL (as a string). If
// the URL cannot be parsed, an error is returned. The HTTPClient is set to
// an empty http.Client, but this can be changed programmatically by setting
// client.HTTPClient. The user can also programmatically set the URL as a
// *url.URL.
func NewClient(urlString string) (*Client, error) {
	if len(urlString) == 0 {
		return nil, fmt.Errorf("client: missing url")
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	token := os.Getenv("ATLAS_TOKEN")
	if token != "" {
		log.Printf("[DEBUG] using ATLAS_TOKEN (%s)", maskString(token))
	}

	client := &Client{
		URL:           parsedURL,
		Token:         token,
		DefaultHeader: make(http.Header),
	}

	client.DefaultHeader.Set("User-Agent", userAgent)
	client.DefaultHeader.Set("Content-Type", "application/vnd.api+json")

	if err := client.init(); err != nil {
		return nil, err
	}

	return client, nil
}

// init() sets defaults on the client.
func (c *Client) init() error {
	c.HTTPClient = cleanhttp.DefaultClient()

	tlsConfig := &tls.Config{}
	if os.Getenv(atlasTLSNoVerifyEnvVar) != "" {
		tlsConfig.InsecureSkipVerify = true
	}
	err := rootcerts.ConfigureTLS(tlsConfig, &rootcerts.Config{
		CAFile: os.Getenv(atlasCAFileEnvVar),
		CAPath: os.Getenv(atlasCAPathEnvVar),
	})
	if err != nil {
		return err
	}

	t := cleanhttp.DefaultTransport()
	t.TLSClientConfig = tlsConfig
	c.HTTPClient.Transport = t
	return nil
}
