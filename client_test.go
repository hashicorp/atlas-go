package harmony

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

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

func TestNewClient_setsDefaultHTTPClient(t *testing.T) {
	client, err := NewClient("https://example.com/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(client.HTTPClient, http.DefaultClient) {
		t.Fatalf("expected %q to equal %q", client.HTTPClient, http.DefaultClient)
	}
}

func TestLogin_missingUsername(t *testing.T) {
	client, err := NewClient("https://example.com/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Login("", "")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}

	expected := "client: missing username"
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected %q to contain %q", err.Error(), expected)
	}
}

func TestLogin_missingPassword(t *testing.T) {
	client, err := NewClient("https://example.com/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Login("username", "")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}

	expected := "client: missing password"
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected %q to contain %q", err.Error(), expected)
	}
}

func TestRequest_getsData(t *testing.T) {
	server := newTestHarmonyServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.request("get", "/_status/200")
	if err != nil {
		t.Fatal(err)
	}

	var body bytes.Buffer
	io.Copy(&body, response.Body)

	expected := "Status code: 200"
	if body.String() != expected {
		t.Fatalf("expected %q to equal %q", body.String(), expected)
	}
}

func TestRequest_returnsError(t *testing.T) {
	server := newTestHarmonyServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.request("get", "/_status/404")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}

	expected := "Status code: 404"
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected %q to contain %q", err.Error(), expected)
	}
}
}
