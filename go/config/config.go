package config

import "fmt"

type Config struct {
	Options Options
	Project Project
}

type Options struct {
	Namespaces  bool
	WorkingPath string
	Template    Template
	Management  Management
}

type Template struct {
	File string
	Path string
	Ext  string
}

type Management struct {
	DB DB
}

type Project struct {
	Name   string
	DB     DB
	Schema Schema
}

type Schema struct {
	Name       string
	Url        string
	Version    string
	Namespaces []SchemaNamespace
}

type SchemaNamespace struct {
	Name        string
	ShortName   string
	TablePrefix string
	Folder      string
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
