package atlas

import (
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultClient_url(t *testing.T) {
	client := DefaultClient()

	if client.URL.String() != atlasDefaultEndpoint {
		t.Fatalf("expected %q to be %q", client.URL.String(), atlasDefaultEndpoint)
	}
}

func TestDefaultClient_urlFromEnvVar(t *testing.T) {
	defer os.Setenv(atlasEndpointEnvVar, os.Getenv(atlasEndpointEnvVar))
	otherEndpoint := "http://127.0.0.1:1234"

	err := os.Setenv(atlasEndpointEnvVar, otherEndpoint)
	if err != nil {
		t.Fatal(err)
	}

	client := DefaultClient()

	if client.URL.String() != otherEndpoint {
		t.Fatalf("expected %q to be %q", client.URL.String(), otherEndpoint)
	}
}

func TestNewClient_badURL(t *testing.T) {
	_, err := NewClient("")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}

	expected := "client: missing url"
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected %q to contain %q", err.Error(), expected)
	}
}

func TestNewClient_parsesURL(t *testing.T) {
	client, err := NewClient("https://example.com/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	expected := &url.URL{
		Scheme: "https",
		Host:   "example.com",
		Path:   "/foo/bar",
	}
	if !reflect.DeepEqual(client.URL, expected) {
		t.Fatalf("expected %q to equal %q", client.URL, expected)
	}
}

func TestNewClient_TLSVerify(t *testing.T) {
	client, err := NewClient("https://example.com/foo/bar")

	if err != nil {
		t.Fatal(err)
	}
	if client.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify != false {
		t.Fatal("Expected InsecureSkipVerify to be false")
	}
}

func TestNewClient_TLSNoVerify(t *testing.T) {
	os.Setenv("ATLAS_TLS_NOVERIFY", "1")
	client, err := NewClient("https://example.com/foo/bar")

	if err != nil {
		t.Fatal(err)
	}
	if client.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify != true {
		t.Fatal("Expected InsecureSkipVerify to be true when ATLAS_TLS_NOVERIFY is set")
	}
	os.Setenv("ATLAS_TLS_NOVERIFY", "")
}

func TestNewClient_setsDefaultHTTPClient(t *testing.T) {
	_, err := NewClient("https://example.com/foo/bar")
	if err != nil {
		t.Fatal(err)
	}
}
