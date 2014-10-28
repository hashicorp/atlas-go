package harmony

import (
	"fmt"
)

// BuildConfig represents a Packer build configuration.
type BuildConfig struct {
	// Username is the namespace under which the build config lives
	Username string

	// Name is the actual name of the build config, unique in the scope
	// of the username.
	Name string
}

// BuildConfig gets a single build configuration by user and name.
func (c *Client) BuildConfig(user, name string) (*BuildConfig, error) {
	endpoint := fmt.Sprintf("/api/v1/packer/build-configurations/%s/%s", user, name)
	request, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var w bcWrapper
	if err := decodeJSON(response, &w); err != nil {
		return nil, err
	}

	return w.BuildConfig, nil

}

// bcWrapper is the API wrapper since the server wraps the resulting object.
type bcWrapper struct {
	BuildConfig *BuildConfig `json:"build_configuration"`
}
