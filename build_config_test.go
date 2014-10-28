package harmony

import (
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
		Username: "hashicorp",
		Name: "existing",
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%#v", actual)
	}
}
