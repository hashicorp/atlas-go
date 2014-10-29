package harmony

import (
	"bytes"
	"reflect"
	"testing"
)

func TestBuildConfig_fetches(t *testing.T) {
	server := newTestHarmonyServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	actual, err := client.BuildConfig("hashicorp", "existing")
	if err != nil {
		t.Fatal(err)
	}

	expected := &BuildConfig{
		User: "hashicorp",
		Name: "existing",
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%#v", actual)
	}
}

func TestCreateBuildConfig(t *testing.T) {
	server := newTestHarmonyServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	err = client.CreateBuildConfig("hashicorp", "existing")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUploadBuildConfigVersion(t *testing.T) {
	server := newTestHarmonyServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	data := new(bytes.Buffer)
	err = client.UploadBuildConfigVersion(&BuildConfigVersion{
		User: "hashicorp",
		Name: "existing",
		Builds: []BuildConfigBuild{
			BuildConfigBuild{Name: "foo", Type: "ami"},
		},
	}, data)
	if err != nil {
		t.Fatal(err)
	}
}
