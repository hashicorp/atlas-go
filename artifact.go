package atlas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

// Artifact represents a single instance of an artifact.
type Artifact struct {
	// User and name are self-explanatory. Tag is the combination
	// of both into "username/name"
	User string `json:"username"`
	Name string `json:"name"`
	Tag  string `json:",omitempty"`
}

// ArtifactVersion represents a single version of an artifact.
type ArtifactVersion struct {
	User     string            `json:"username"`
	Name     string            `json:"name"`
	Tag      string            `json:",omitempty"`
	Type     string            `json:"artifact_type"`
	ID       string            `json:"external_id"`
	Version  int               `json:"version"`
	Metadata map[string]string `json:"metadata"`
	File     bool              `json:"file"`

	UploadPath  string `json:"upload_path"`
	UploadToken string `json:"upload_token"`
}

// ArtifactSearchOpts are the options used to search for an artifact.
type ArtifactSearchOpts struct {
	User string
	Name string
	Type string

	Version  string
	Metadata map[string]string
}

// UploadArtifactOpts are the options used to upload an artifact.
type UploadArtifactOpts struct {
	User     string
	Name     string
	Type     string
	ID       string
	File     io.Reader
	Metadata map[string]string
	BuildId  int
}

func (o *UploadArtifactOpts) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"artifact_version": map[string]interface{}{
			"id":       o.ID,
			"file":     o.File != nil,
			"metadata": o.Metadata,
			"build_id": o.BuildId,
		},
	})
}

// This is the value that should be used for metadata in ArtifactSearchOpts
// if you don't care what the value is.
const MetadataAnyValue = "943febbf-589f-401b-8f25-58f6d8786848"

func (c *Client) Artifact(user, name string) (*Artifact, error) {
	endpoint := fmt.Sprintf("/api/v1/artifacts/%s/%s", user, name)
	request, err := c.Request("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var aw artifactWrapper
	if err := decodeJSON(response, &aw); err != nil {
		return nil, err
	}

	return aw.Artifact, nil
}

func (c *Client) ArtifactSearch(opts *ArtifactSearchOpts) ([]*ArtifactVersion, error) {
	params := make(map[string]string)
	if opts.Version != "" {
		params["version"] = opts.Version
	}

	i := 1
	for k, v := range opts.Metadata {
		prefix := fmt.Sprintf("metadata.%d.", i)
		params[prefix+"key"] = k
		if v != MetadataAnyValue {
			params[prefix+"value"] = v
		}

		i++
	}

	endpoint := fmt.Sprintf(
		"/api/v1/artifacts/%s/%s/%s/search",
		opts.User, opts.Name, opts.Type)
	request, err := c.Request("GET", endpoint, &RequestOptions{
		Params: params,
	})
	if err != nil {
		return nil, err
	}

	response, err := checkResp(c.HTTPClient.Do(request))
	if err != nil {
		return nil, err
	}

	var w artifactSearchWrapper
	if err := decodeJSON(response, &w); err != nil {
		return nil, err
	}

	return w.Versions, nil
}

func (c *Client) CreateArtifact(user, name string) (*Artifact, error) {
	body, err := json.Marshal(&artifactWrapper{&Artifact{
		User: user,
		Name: name,
	}})
	if err != nil {
		return nil, err
	}

	endpoint := "/api/v1/artifacts"
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

	var aw artifactWrapper
	if err := decodeJSON(response, &aw); err != nil {
		return nil, err
	}

	return aw.Artifact, nil
}

func (c *Client) ArtifactFileURL(av *ArtifactVersion) (*url.URL, error) {
	if !av.File {
		return nil, nil
	}

	u := *c.URL
	u.Path = fmt.Sprintf("/api/v1/artifacts/%s/%s/%s/file",
		av.User, av.Name, av.Type)
	return &u, nil
}

func (c *Client) UploadArtifact(opts *UploadArtifactOpts) (*ArtifactVersion, error) {
	endpoint := fmt.Sprintf("/api/v1/artifacts/%s/%s/%s",
		opts.User, opts.Name, opts.Type)

	body, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

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

	var av ArtifactVersion
	if err := decodeJSON(response, &av); err != nil {
		return nil, err
	}

	if opts.File != nil {
		if err := c.putFile(av.UploadPath, opts.File); err != nil {
			return nil, err
		}
	}

	return &av, nil
}

type artifactWrapper struct {
	Artifact *Artifact `json:"artifact"`
}

type artifactSearchWrapper struct {
	Versions []*ArtifactVersion
}

type artifactVersionWrapper struct {
	Version *ArtifactVersion
}
