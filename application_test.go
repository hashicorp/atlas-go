package harmony

import (
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
