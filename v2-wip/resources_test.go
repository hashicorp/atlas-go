package atlas

import (
	"reflect"
	"testing"
)

func TestRequest_stateIncludedVersions(t *testing.T) {
	server := newTestAtlasServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	request, err := client.RawRequest("GET", "/api/v2/states/150-include-versions", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := checkResp(client.HTTPClient.Do(request))
	if err != nil {
		t.Fatal(err)
	}

	var data State
	if err := Unmarshal(resp.Body, &data); err != nil {
		t.Fatal(err)
	}

	expected := State{
		ID:        150,
		CreatedAt: "2017-03-10T00:51:19.359Z",
		UpdatedAt: "2017-03-10T00:52:44.233Z",
		Versions: []*StateVersion{
			&StateVersion{
				ID:        424,
				CreatedAt: "2017-03-10T00:52:44.225Z",
				UpdatedAt: "2017-03-10T00:52:44.545Z",
				Version:   1,
				Serial:    3,
				Tfstate:   `{"version": 3}`,
			},
		},
	}
	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("bad:\n\n%#v\n\n%#v", data, expected)
	}
}

func TestRequest_stateVersion(t *testing.T) {
	server := newTestAtlasServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	request, err := client.RawRequest("GET", "/api/v2/state-versions/424", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := checkResp(client.HTTPClient.Do(request))
	if err != nil {
		t.Fatal(err)
	}

	var data StateVersion
	if err := Unmarshal(resp.Body, &data); err != nil {
		t.Fatal(err)
	}

	expected := StateVersion{
		ID:        424,
		CreatedAt: "2017-03-10T00:52:44.225Z",
		UpdatedAt: "2017-03-10T00:52:44.545Z",
		Version:   1,
		Serial:    3,
		Tfstate:   `{"version": 3}`,
	}
	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("bad:\n\n%#v\n\n%#v", data, expected)
	}
}
