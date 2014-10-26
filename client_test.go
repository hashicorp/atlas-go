package harmony

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

	expected := "client: misisng url"
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

func TestNewClient_setsHTTPClient(t *testing.T) {
	client, err := NewClient("https://example.com/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(client.httpClient, http.DefaultClient) {
		t.Fatalf("expected %q to equal %q", client.httpClient, http.DefaultClient)
	}
}

func TestSetHTTPClient_setsHTTPClient(t *testing.T) {
	client, err := NewClient("https://example.com/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	expected := &http.Client{}
	client.SetHTTPClient(expected)

	if !reflect.DeepEqual(client.httpClient, expected) {
		t.Fatalf("expected %q to equal %q", client.httpClient, expected)
	}
}

func TestSetToken_setsToken(t *testing.T) {
	client, err := NewClient("https://example.com/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	expected := "token"
	client.SetToken(expected)

	if !reflect.DeepEqual(client.token, expected) {
		t.Fatalf("expected %q to equal %q", client.token, expected)
	}
}

func TestRequest_temp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"text": "testing"}`)
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.request("get", "/zip")
	if err != nil {
		t.Fatal(err)
	}

	_ = response
}
