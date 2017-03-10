package atlas

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestMarshal_omitID(t *testing.T) {
	data := StateVersion{
		Version: 1,
		Serial:  3,
		Tfstate: `{"version": 3}`,
	}

	var buf bytes.Buffer
	if err := Marshal(&buf, &data); err != nil {
		t.Fatal(err)
	}

	var actual map[string]interface{}
	if err := json.NewDecoder(&buf).Decode(&actual); err != nil {
		t.Fatal(err)
	}

	t.Logf("Actual: %#v", actual)

	actualData := actual["data"].(map[string]interface{})
	if _, ok := actualData["id"]; ok {
		t.Fatal("should not have ID")
	}
}

func TestMarshal_includeSetID(t *testing.T) {
	data := StateVersion{
		ID:      150,
		Version: 1,
		Serial:  3,
		Tfstate: `{"version": 3}`,
	}

	var buf bytes.Buffer
	if err := Marshal(&buf, &data); err != nil {
		t.Fatal(err)
	}

	var actual map[string]interface{}
	if err := json.NewDecoder(&buf).Decode(&actual); err != nil {
		t.Fatal(err)
	}

	t.Logf("Actual: %#v", actual)

	actualData := actual["data"].(map[string]interface{})
	if _, ok := actualData["id"]; !ok {
		t.Fatal("should have ID")
	}
}
