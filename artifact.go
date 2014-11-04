package harmony

import (
	"fmt"
)

// Artifact represents a single instance of an artifact.
type Artifact struct {
	// User and name are self-explanatory. Tag is the combination
	// of both into "username/name"
	User string `json:"username"`
	Name string
	Tag  string
}

// ArtifactVersion represents a single version of an artifact.
type ArtifactVersion struct {
	Type     string `json:"artifact_type"`
	ID       string `json:"external_id"`
	Version  int
	Metadata map[string]string
	File     bool

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

type artifactWrapper struct {
	Artifact *Artifact
}

type artifactSearchWrapper struct {
	Versions []*ArtifactVersion
}
