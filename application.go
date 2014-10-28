package harmony

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Application represents a single instance of an Application on the Harmony
// server.
type Application struct {
	// User is the namespace (username or organization) under which the
	// Application resides
	User string `json:"user"`

	// Name is the name of the Application
	Name string `json:"name"`
}

// Application gets the Application by the given user space and name. In the
// event the Application is not found, or for any other non-200 responses, an
// error is returned.
func (c *Client) Application(user, name string) (*Application, error) {
	endpoint := fmt.Sprintf("/api/v2/vagrant/applications/%s/%s", user, name)
	request, err := c.NewRequest("HEAD", endpoint, nil)
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var app Application
	if err := decodeJSON(response, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

// CreateApplication creates a new Application under the given user with
// the given name. If the Application is created successfully, it is returned.
// If the server returns any errors, an error is returned.
func (c *Client) CreateApplication(user, name string) (*Application, error) {
	body, err := json.Marshal(&Application{
		User: user,
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	endpoint := "/api/v2/vagrant/applications"
	request, err := c.NewRequest("POST", endpoint, &RequestOptions{
		Body: bytes.NewBuffer(body),
	})
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var app Application
	if err := decodeJSON(response, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

// ApplicationVersion represents a specific version of an Application in
// Harmony. It is actually an upload container/wrapper.
type ApplicationVersion struct {
	UploadPath string `json:"upload_path"`
	Token      string `json:"token"`
}

// CreateVersion makes a new ApplicationVersion for the Application. There are
// no parameters to this method. If the server is unable to create a new
// version, an error is returned.
func (c *Client) CreateApplicationVersion(app *Application) (*ApplicationVersion, error) {
	endpoint := fmt.Sprintf("/api/v2/vagrant/applications/%s/%s/version",
		app.User, app.Name)

	request, err := c.NewRequest("POST", endpoint, nil)
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var av ApplicationVersion
	if err := decodeJSON(response, &av); err != nil {
		return nil, err
	}

	return &av, nil
}

// Upload accepts data as an io.Reader and PUTs data to the ApplicationVersion's
// UploadPath. If any errors occur before or during the upload, they are
// returned.
func (av *ApplicationVersion) Upload(data io.Reader) error {
	client, err := NewClient(av.UploadPath)
	if err != nil {
		return err
	}

	request, err := client.NewRequest("PUT", "/", &RequestOptions{
		Body: data,
	})
	if err != nil {
		return err
	}

	response, err := checkResp(client.HTTPClient.Do(request))
	if err != nil {
		return err
	}

	_ = response // TODO

	return nil
}
