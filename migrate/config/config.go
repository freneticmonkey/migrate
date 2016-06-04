package config

import "fmt"

type Config struct {
	Options
	Project
}

type Options struct {
	WorkingPath string
	Management
}

type Management struct {
	DB
}

type Project struct {
	Name string
	DB
	Schema
	LocalSchema
}

type Schema struct {
	Name    string
	Url     string
	Version string
	Folders []string
}

type LocalSchema struct {
	Path string
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
