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
	CreatedAt     string          `jsonapi:"attr,created-at"`
	UpdatedAt     string          `jsonapi:"attr,updated-at"`
	Configuration *Configuration  `jsonapi:"relation,configuration"`
	Environment   *Environment    `jsonapi:"relation,environment"`
	Versions      []*StateVersion `jsonapi:"relation,versions"`
	HeadVersion   *StateVersion   `jsonapi:"relation,head-version"`
}

type StateVersion struct {
	ID        int    `jsonapi:"primary,state-versions"`
	CreatedAt string `jsonapi:"attr,created-at"`
	UpdatedAt string `jsonapi:"attr,updated-at"`
	Version   int    `jsonapi:"attr,version"`
	Serial    int    `jsonapi:"attr,serial"`
	Tfstate   string `jsonapi:"attr,tfstate-file"`
	State     *State `jsonapi:"relation,state"`
}
