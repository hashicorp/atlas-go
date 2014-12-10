package atlas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// appWrapper is the API wrapper since the server wraps the resulting object.
type appWrapper struct {
	Application *App `json:"application"`
}

// App represents a single instance of an application on the Atlas server.
type App struct {
	// User is the namespace (username or organization) under which the
	// Atlas application resides
	User string `json:"username"`

	// Name is the name of the application
	Name string `json:"name"`
}

// Slug returns the slug format for this App (User/Name)
func (a *App) Slug() string {
	return fmt.Sprintf("%s/%s", a.User, a.Name)
}

// App gets the App by the given user space and name. In the event the App is
// not found (404), or for any other non-200 responses, an error is returned.
func (c *Client) App(user, name string) (*App, error) {
	endpoint := fmt.Sprintf("/api/v1/vagrant/applications/%s/%s", user, name)
	request, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var app App
	if err := decodeJSON(response, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

// CreateApp creates a new App under the given user with the given name. If the
// App is created successfully, it is returned. If the server returns any
// errors, an error is returned.
func (c *Client) CreateApp(user, name string) (*App, error) {
	body, err := json.Marshal(&appWrapper{&App{
		User: user,
		Name: name,
	}})
	if err != nil {
		return nil, err
	}

	endpoint := "/api/v1/vagrant/applications"
	request, err := c.Request("POST", endpoint, &RequestOptions{
		Body: bytes.NewReader(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	})
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var app App
	if err := decodeJSON(response, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

// appVersion represents a specific version of an App in Atlas. It is actually
// an upload container/wrapper.
type appVersion struct {
	UploadPath string `json:"upload_path"`
	Token      string `json:"token"`
	Version    uint64 `json:"version"`
}

// UploadApp creates and uploads a new version for the App. If the server does not
// find the application, an error is returned. If the server does not accept the
// data, an error is returned.
//
// It is the responsibility of the caller to create a properly-formed data
// object; this method blindly passes along the contents of the io.Reader.
func (c *Client) UploadApp(app *App, data io.Reader, size int64) (uint64, error) {
	endpoint := fmt.Sprintf("/api/v1/vagrant/applications/%s/%s/versions",
		app.User, app.Name)

	request, err := c.Request("POST", endpoint, nil)
	if err != nil {
		return 0, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return 0, err
	}

	var av appVersion
	if err := decodeJSON(response, &av); err != nil {
		return 0, err
	}

	if err := c.putFile(av.UploadPath, data, size); err != nil {
		return 0, err
	}

	return av.Version, nil
}
