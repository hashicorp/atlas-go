package harmony

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// AppWraper
type AppWrapper struct {
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

// App gets the App by the given user space and name. In the event the App is
// not found (404), or for any other non-200 responses, an error is returned.
func (c *Client) App(user, name string) (*App, error) {
	endpoint := fmt.Sprintf("/api/v2/vagrant/applications/%s/%s", user, name)
	request, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var aw AppWrapper
	if err := decodeJSON(response, &aw); err != nil {
		return nil, err
	}

	return aw.Application, nil
}

// CreateApp creates a new App under the given user with the given name. If the
// App is created successfully, it is returned. If the server returns any
// errors, an error is returned.
func (c *Client) CreateApp(user, name string) (*App, error) {
	body, err := json.Marshal(&AppWrapper{&App{
		User: user,
		Name: name,
	}})
	if err != nil {
		return nil, err
	}

	endpoint := "/api/v2/vagrant/applications"
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

	var aw AppWrapper
	if err := decodeJSON(response, &aw); err != nil {
		return nil, err
	}

	return aw.Application, nil
}

// AppVersion represents a specific version of an App in Harmony. It is actually
// an upload container/wrapper.
type AppVersion struct {
	UploadPath string `json:"upload_path"`
	Token      string `json:"token"`
}

// CreateVersion makes a new AppVersion for the App. If the server is unable to
// create a new version, an error is returned.
func (c *Client) CreateAppVersion(app *App) (*AppVersion, error) {
	endpoint := fmt.Sprintf("/api/v2/vagrant/applications/%s/%s/version",
		app.User, app.Name)

	request, err := c.Request("POST", endpoint, nil)
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var av AppVersion
	if err := decodeJSON(response, &av); err != nil {
		return nil, err
	}

	return &av, nil
}

// UploadAppVersion accepts data as an io.Reader and PUTs data to the
// AppVersion's UploadPath. If any errors occur before or during the upload,
// they are returned.
func (client *Client) UploadAppVersion(av *AppVersion, data io.Reader) error {
	u, err := url.Parse(av.UploadPath)
	if err != nil {
		return err
	}

	// Use the private rawRequest function here to avoid appending the
	// access_token and being restricted to the Harmony namespace, since binstore
	// lives under a different root URL.
	request, err := client.rawRequest("PUT", u, &RequestOptions{
		Body: data,
	})
	if err != nil {
		return err
	}

	if _, err := checkResp(client.HTTPClient.Do(request)); err != nil {
		return err
	}

	return nil
}
