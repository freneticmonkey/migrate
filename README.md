# Migrate

# Overview
This tool is intended to be a very simple database migration system which migrates schema by comparing the output of mysql 'show create tables' to a simple YAML table definition.  The differences are then converted into MySQL ALTER TABLE statements which can be applied to the target database.  The YAML schema is expected to be version controlled using a Git repository, which also is the source of the migration versioning information and used to validate schema state.

## Management
Migrations are tracked and approval managed using a separate database.  

Each supported property of a table is assigned a unique identifier which is stored within the management database.  This allows non-intrusive management of the schema leaving the target schema free of tags or embedded identifiers.

## Migration execution
Migrate uses the built in Go MySQL database driver for simple operations such as CREATE/DROP TABLE, with the option to use the pt-online-schema-change tool developed by Percona for long running operations, without requiring the downtime for the database.

For quick introduction on how to use Migrate see the [CLI](docs/CLI.md) and [Getting Started](docs/GETTING_STARTED.md) docs.

## REST API

Migrate can also run as a REST API service which allows for schema management via a REST API.  For more info see the [REST API](docs/RESTAPI.md) docs.
