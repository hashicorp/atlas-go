package atlas

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRequest_getsData(t *testing.T) {
	server := newTestAtlasServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	request, err := client.RawRequest("GET", "/_status/200", nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := checkResp(client.HTTPClient.Do(request)); err != nil {
		t.Fatal(err)
	}
}

func TestRequest_error(t *testing.T) {
	server := newTestAtlasServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	request, err := client.RawRequest("GET", "/_error", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = checkResp(client.HTTPClient.Do(request))
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}

	expected := &ErrorDoc{
		Errors: []*Error{
			&Error{
				Title:  "bad",
				Detail: "detail",
				Code:   "114",
				Status: "400",
			},
		},
	}

	if !reflect.DeepEqual(err, expected) {
		t.Fatalf("expected %+v to be %+v", err, expected)
	}
}

func TestRequest_notFoundError(t *testing.T) {
	server := newTestAtlasServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	request, err := client.RawRequest("GET", "/_status/404", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = checkResp(client.HTTPClient.Do(request))
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}

	if err != ErrNotFound {
		t.Fatalf("bad error: %#v", err)
	}
}

func TestRequestJSON_decodesData(t *testing.T) {
	server := newTestAtlasServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	request, err := client.RawRequest("GET", "/_json", nil)
	if err != nil {
		t.Fatal(err)
	}

	response, err := checkResp(client.HTTPClient.Do(request))
	if err != nil {
		t.Fatal(err)
	}

	var decoded struct{ Ok bool }
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		t.Fatal(err)
	}

	if !decoded.Ok {
		t.Fatal("expected decoded response to be Ok, but was not")
	}
}

// check that our DefaultHeader works correctly, along with it providing
// User-Agent
func TestClient_defaultHeaders(t *testing.T) {
	server := newTestAtlasServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	testHeader := "Atlas-Test"
	testHeaderVal := "default header test"
	client.DefaultHeader.Set(testHeader, testHeaderVal)

	request, err := client.RawRequest("GET", "/_test", nil)
	if err != nil {
		t.Fatal(err)
	}

	response, err := checkResp(client.HTTPClient.Do(request))
	if err != nil {
		t.Fatal(err)
	}

	decoded := &clientTestResp{}
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		t.Fatal(err)
	}

	// Make sure User-Agent is set correctly
	if decoded.Header.Get("User-Agent") != userAgent {
		t.Fatal("User-Agent reported as", decoded.Header.Get("User-Agent"))
	}

	// look for our test header
	if decoded.Header.Get(testHeader) != testHeaderVal {
		t.Fatalf("DefaultHeader %q reported as %q", testHeader, testHeaderVal)
	}
}
