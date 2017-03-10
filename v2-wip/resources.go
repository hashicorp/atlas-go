package atlas

// This file contains the JSON API resource types for the V2 API. The
// documentation for each of these resources can be found in the official
// API documentation and won't be reproduced here.

type Organization struct {
	ID             int              `jsonapi:"primary,organizations"`
	Username       string           `jsonapi:"attr,username"`
	Email          string           `jsonapi:"attr,email"`
	Configurations []*Configuration `jsonapi:"relation,configurations"`
	Environments   []*Environment   `jsonapi:"relation,environments"`
}

type Configuration struct {
	ID           int                     `jsonapi:"primary,configurations"`
	Name         string                  `jsonapi:"attr,name"`
	Organization []*Organization         `jsonapi:"relation,organization"`
	Versions     []*ConfigurationVersion `jsonapi:"relation,versions"`
}

type ConfigurationVersion struct {
	ID            int               `jsonapi:"primary,configuration-versions"`
	Version       int               `jsonapi:"attr,version"`
	Hidden        bool              `jsonapi:"attr,is-hidden"`
	Metadata      map[string]string `jsonapi:"attr,metadata"`
	Status        string            `jsonapi:"attr,status"`
	Configuration []*Configuration  `jsonapi:"relation,configuration"`
}

type Environment struct {
	// TODO
}

type State struct {
	ID            int             `jsonapi:"primary,states"`
	CreatedAt     string          `jsonapi:"attr,created-at,omitempty"`
	UpdatedAt     string          `jsonapi:"attr,updated-at,omitempty"`
	Configuration *Configuration  `jsonapi:"relation,configuration,omitempty"`
	Environment   *Environment    `jsonapi:"relation,environment,omitempty"`
	Versions      []*StateVersion `jsonapi:"relation,versions,omitempty"`
	HeadVersion   *StateVersion   `jsonapi:"relation,head-version,omitempty"`
}

type StateVersion struct {
	ID        int    `jsonapi:"primary,state-versions,omitempty"`
	CreatedAt string `jsonapi:"attr,created-at,omitempty"`
	UpdatedAt string `jsonapi:"attr,updated-at,omitempty"`
	Version   int    `jsonapi:"attr,version,omitempty"`
	Serial    int    `jsonapi:"attr,serial,omitempty"`
	Tfstate   string `jsonapi:"attr,tfstate-file,omitempty"`
	State     *State `jsonapi:"relation,state,omitempty"`
}
