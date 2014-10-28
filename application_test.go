package harmony

import (
	"bytes"
	"path"
	"testing"
)

func TestApp_fetchesApp(t *testing.T) {
	server := newTestHarmonyServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	app, err := client.App("hashicorp", "existing")
	if err != nil {
		t.Fatal(err)
	}

	if app.User != "hashicorp" {
		t.Errorf("expected %q to be %q", app.User, "hashicorp")
	}

	if app.Name != "existing" {
		t.Errorf("expected %q to be %q", app.Name, "existing")
	}
}

func TestApp_returnsErrorNoApp(t *testing.T) {
	server := newTestHarmonyServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.App("hashicorp", "newproject")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}
}

func TestCreateApp_createsAndReturnsApp(t *testing.T) {
	server := newTestHarmonyServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	app, err := client.CreateApp("hashicorp", "newproject")
	if err != nil {
		t.Fatal(err)
	}

	if app.User != "hashicorp" {
		t.Errorf("expected %q to be %q", app.User, "hashicorp")
	}

	if app.Name != "newproject" {
		t.Errorf("expected %q to be %q", app.Name, "newproject")
	}
}

func TestCreateApp_returnsErrorExistingApp(t *testing.T) {
	server := newTestHarmonyServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.CreateApp("hashicorp", "existing")
	if err == nil {
		t.Fatal("expected error, but nothing was returned")
	}
}

func TestCreateAppVersion_createsAndReturnsVersion(t *testing.T) {
	server := newTestHarmonyServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	av, err := client.CreateAppVersion(&App{
		User: "hashicorp",
		Name: "existing",
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := "https://binstore.hashicorp.com/630e42d9-2364-2412-4121-18266770468e"
	if av.UploadPath != expected {
		t.Errorf("expected %q to be %q", av.UploadPath, expected)
	}

	expected = "630e42d9-2364-2412-4121-18266770468e"
	if av.Token != expected {
		t.Errorf("expected %q to be %q", av.Token, expected)
	}
}

func TestUploadAppVersion_createsVersion(t *testing.T) {
	server := newTestHarmonyServer(t)
	defer server.Stop()

	client, err := NewClient(server.URL.String())
	if err != nil {
		t.Fatal(err)
	}

	buff := new(bytes.Buffer)
	u := *server.URL
	u.Path = path.Join(server.URL.Path, "_binstore")
	av := &AppVersion{
		UploadPath: u.String(),
		Token:      "abcd-1234",
	}
	err = client.UploadAppVersion(av, buff)
	if err != nil {
		t.Fatal(err)
	}
}
