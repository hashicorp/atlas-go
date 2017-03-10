package atlas

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/google/jsonapi"
)

// Unmarshal will unmarshal a successful API request into the given structure.
// The structure is expected to be one of the structures in resources.go.
//
// This is simply a wrapper around the google/jsonapi library for now. We
// provide these wrappers to sometimes add custom behavior (such as Meta
// decoding).
func Unmarshal(r io.Reader, m interface{}) error {
	return jsonapi.UnmarshalPayload(r, m)
}

// Marshal will marshal a single resource structure for making a request.
//
// This resolves an issue with the standard google/jsonapi library by
// excluding specific fields that are either unnecessary or strictly
// prohibited by the jsonapi specification. For example, the jsonapi library
// encodes "id" even if it is zero, but jsonapi requires this be omitted
// if it isn't supposed to be specified. This marshal function handles that.
func Marshal(w io.Writer, m interface{}) error {
	payload, err := jsonapi.MarshalOne(m)
	if err != nil {
		return err
	}

	// We don't want to send any included, since the TFE API doesn't use it.
	payload.Included = nil

	// Marshal to an in-memory buffer first to allow us to do replacement.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}

	// Delete the ID from the data
	var temp map[string]interface{}
	if err := json.NewDecoder(&buf).Decode(&temp); err != nil {
		return err
	}
	tempData := temp["data"].(map[string]interface{})
	delete(tempData, "id")

	// Re-encode
	return json.NewEncoder(w).Encode(temp)
}
