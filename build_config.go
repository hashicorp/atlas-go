package harmony

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// BuildConfig represents a Packer build configuration.
type BuildConfig struct {
	// Username is the namespace under which the build config lives
	Username string

	// Name is the actual name of the build config, unique in the scope
	// of the username.
	Name string
}

// BuildConfigVersion represents a single uploaded (or uploadable) version
// of a build configuration.
type BuildConfigVersion struct {
	// The fields below are the username/name combo to uniquely identify
	// a build config.
	Username string
	Name     string

	// Builds is the list of builds that this version supports.
	Builds []BuildConfigBuild
}

// BuildConfigBuild is a single build that is present in an uploaded
// build configuration.
type BuildConfigBuild struct {
	// Name is a unique name for this build
	Name string

	// Type is the type of builder that this build needs to run on,
	// such as "amazon-ebs" or "qemu".
	Type string
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

// CreateBuildConfigVersion creates a single build configuration version
// and uploads the template associated with it.
func (c *Client) CreateBuildConfigVersion(v *BuildConfigVersion, tpl io.Reader) error {
	endpoint := fmt.Sprintf("/api/v1/packer/build-configurations/%s/%s/version",
		v.Username, v.Name)

	var bodyData bcCreateWrapper
	bodyData.Version.Builds = v.Builds
	body, err := json.Marshal(bodyData)
	if err != nil {
		return err
	}

	request, err := c.Request("POST", endpoint, &RequestOptions{
		Body: bytes.NewBuffer(body),
	})
	if err != nil {
		return err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return err
	}

	var data bcCreate
	if err := decodeJSON(response, &data); err != nil {
		return err
	}

	uploadUrl, err := url.Parse(data.UploadPath)
	if err != nil {
		return err
	}

	// Use the private rawRequest function here to avoid appending the
	// access_token and being restricted to the Harmony namespace, since binstore
	// lives under a different root URL.
	request, err = c.rawRequest("PUT", uploadUrl, &RequestOptions{
		Body: tpl,
	})
	if err != nil {
		return err
	}

	if _, err := checkResp(c.HTTPClient.Do(request)); err != nil {
		return err
	}

	return nil
}

// bcWrapper is the API wrapper since the server wraps the resulting object.
type bcWrapper struct {
	BuildConfig *BuildConfig `json:"build_configuration"`
}

// bcCreate is the struct returned when creating a build configuration.
type bcCreate struct {
	UploadPath string `json:"upload_path"`
}

// bcCreateWrapper is the wrapper for creating a build config.
type bcCreateWrapper struct {
	Version struct {
		Builds []BuildConfigBuild
	}
}
