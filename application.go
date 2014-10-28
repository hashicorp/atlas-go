package harmony

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// AppWraper
type appWrapper struct {
	Application *App `json:"application"`
}

// App represents a single instance of an application on the Harmony server.
type App struct {
	// User is the namespace (username or organization) under which the
	// Harmony application resides
	User string `json:"user"`

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

	var aw appWrapper
	if err := decodeJSON(response, &aw); err != nil {
		return nil, err
	}

	return aw.Application, nil
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
		Body: bytes.NewBuffer(body),
	})
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var aw appWrapper
	if err := decodeJSON(response, &aw); err != nil {
		return nil, err
	}

	return aw.Application, nil
}

// appVersion represents a specific version of an App in Harmony. It is actually
// an upload container/wrapper.
type appVersion struct {
	UploadPath *url.URL `json:"upload_path"`
	Token      string   `json:"token"`
	Version    uint64   `json:"version"`
}

// UploadApp creates and uploads a new version for the App. If the server is
// find the application, an error is returned. If the server does not accept the
// data, an error is returned.
//
// It is the responsibility of the caller to create a properly-formed data
// object; this method blindly passes along the contents of the io.Reader.
func (c *Client) UploadApp(app *App, data io.Reader) error {
	endpoint := fmt.Sprintf("/api/v1/vagrant/applications/%s/%s/version",
		app.User, app.Name)

	request, err := c.Request("POST", endpoint, nil)
	if err != nil {
		return err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return err
	}

	var av appVersion
	if err := decodeJSON(response, &av); err != nil {
		return err
	}

	// Use the private rawRequest function here to avoid appending the
	// access_token and being restricted to the Harmony namespace, since binstore
	// lives under a different root URL.
	request, err = c.rawRequest("PUT", av.UploadPath, &RequestOptions{
		Body: data,
	})
	if err != nil {
		return err
	}

	if _, err := checkResp(c.HTTPClient.Do(request)); err != nil {
		return err
	}

	return nil
}
