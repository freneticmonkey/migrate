package config

import "fmt"

type Config struct {
	Options Options
	Project Project
}

type Options struct {
	Namespaces  bool
	WorkingPath string
	Management  Management
	GrayLog     GrayLog
}

type Generation struct {
	Templates []Template
}

type Template struct {
	Name       string
	File       string
	Path       string
	FileFormat string
	Ext        string
}

type GrayLog struct {
	Hostname        string
	Port            int
	Connection      string
	MaxChunkSizeWan int
	MaxChunkSizeLan int
	Parameters      []GrayLogParameter
}

type GrayLogParameter struct {
	Name  string
	Value string
}

type Management struct {
	DB DB
}

type Project struct {
	Name       string
	DB         DB
	Generation Generation
	Schema 	   Schema
	Git        Git
}

type Schema struct {
	Namespaces []SchemaNamespace
}

type Git struct {
	Name       string
	Url        string
	Version    string
}

type SchemaNamespace struct {
	Name        string
	TablePrefix string
	Path        string
}

type DB struct {
	Username    string
	Password    string
	Ip          string
	Port        int
	Database    string
	Environment string
}

func (db DB) ConnectString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", db.Username, db.Password, db.Ip, db.Port, db.Database)
}
