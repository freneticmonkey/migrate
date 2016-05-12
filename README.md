# Migrate

# Overview
This tool is intended to be a very simple database migration system which migrates schema by comparing the output of mysql 'show create tables' to a simple YAML table definition.  The differences are then converted into MySQL ALTER TABLE statements which can be applied to the target database.  The YAML schema is expected to be version controlled using a Git repository, which also is the source of the migration versioning information and used to validate schema state.

## Management
Migrations are tracked and approval managed using a separate database.  

Each supported property of a table is assigned a unique identifier which is stored within the management database.  This allows non-intrusive management of the schema leaving the target schema pristine.

## Migration execution
The tool uses the built in Go MySQL database driver for simple operations such as CREATE/DROP TABLE, with the option to use the pt-online-schema-change tool developed by Percona for long running operations, without requiring the downtime for the database.

# Feature List

- [x] Simple YAML Config
- [x] Reads YAML Table definitions
- [x] Reads MySQL show create table output
- [X] Management DB
- [ ] Migration Sign Off


# TODO ESSENTIAL
- [X] Compare Table structs
- [X] Generate MySQL (ALTER TABLE/CREATE TABLE, DROP TABLE) statements from diff
- [X] Verify unique ids after deserialisation
- [X] Serialize Table structs to YAML
- [X] Serialize Table structs to MySQL
- [X] Namespace support
- [ ] AUTO_INC support
- [ ] Metadata for ids
- [X] Git repo/path/version/time support
- [ ] Human readable migration output - 'Git diff'
- [ ] Initialise Target Database
- [ ] Initialise Management Database
- [ ] Setup from existing target database
- [X] Provide detailed Schema Validation Errors
- [ ] Support local Git Schema changes (Avoiding Git Clone wiping any uncommitted changes from schema)

# TODO Management
- [ ] Implement tables:
    - [ ] Database
    - [X] Migration
    - [X] MigrationStep
    - [X] Metadata

# TODO Migration
- [ ] Standard Go MySQL Driver Migrations
- [ ] pt-online-schema-change Migrations
- [ ] Validation
    - [ ] Creating old Migrations
    - [ ] Creating empty Migrations
    - [ ] Executing old Migrations
    - [ ] Executing cancelled/denied Migrations

# TODO Web Server
- [ ] REST API
    - [ ] Migration
        - [ ] View
        - [ ] Update status

# TODO TOOLING
- [X] Hash generation for embedded identifiers
