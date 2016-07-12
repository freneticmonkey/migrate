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
- [x] Management DB
- [x] Migration Sign Off


# TODO ESSENTIAL
- [x] Compare Table structs
- [x] Generate MySQL (ALTER TABLE/CREATE TABLE, DROP TABLE) statements from diff
- [x] Verify unique ids after deserialisation
- [x] Serialize Table structs to YAML
- [x] Serialize Table structs to MySQL
- [x] Namespace support
- [x] Metadata for ids
- [x] Git repo/path/version/time support
- [x] Human readable migration output - 'Git diff'
- [x] Initialise Management Database
- [x] Provide detailed Schema Validation Errors
- [x] AUTO_INC support
- [x] Implement migrations
- [x] Implement migration validation checks
- [x] Add flag to allow for destructive migrations (drop table, rename column)
- [x] Initialise Project Database
- [x] Setup from existing project database
- [x] Support local Git Schema changes (Avoiding Git Clone wiping any uncommitted changes from schema)
- [x] Load Config via URL
- [x] Add environment setting support

# TODO Management
- [x] Implement tables:
    - [x] TargetDatabase
    - [x] Migration
    - [x] MigrationStep
    - [x] Metadata

# TODO Utilities
- [x] Init management database
- [x] Init project DB
    - [x] Existing
        - [x] Generate Ids
        - [x] YAML Schema
        - [x] Register Metadata to Mgmt DB
    - [x] New
        - [x] Sandbox

# TODO Migration
- [x] Standard Go MySQL Driver Migrations
- [x] pt-online-schema-change Migrations (not tested)
- [x] Validation
    - [x] Creating old Migrations
    - [x] Creating duplicate Migrations
    - [x] Creating empty Migrations
    - [x] Executing old Migrations
    - [x] Executing cancelled/denied Migrations

# TODO REST API Web Server
- [x] REST API
    - [x] Database
        - [x] View
    - [x] Migration
        - [x] View
        - [x] Change status
    - [x] Migration Step
        - [x] View
        - [x] Change status
- [x] Static Serving
- [ ] REST API Front end - TODO => github.com/freneticmonkey/migrate-ui

# TODO TOOLING
- [x] Hash generation for embedded identifiers


# TODO Tests
- [x] MySQL Parsing
- [x] YAML Parsing
- [ ] Difference Engine
    - [ ] Difference detection
    - [ ] MySQL Statement Generation
- [ ] Create Migration
    - [ ] MySQL Statement Validity
    - [ ] Git validation
        - [ ] Can create new
        - [ ] Can't create old
    - [ ] Migration correctly stored in database
- [ ] Run Migration
    - [ ] Successfully applied
    - [ ] Approval
        - [ ] Will run approved
        - [ ] Won't run denied
    - [ ] Ensure single migration execution

# TODO Future / Maybe
- [ ] Add strict mode which hashes the create table statement and stores it for validation against each of the migration_steps.
- [ ] Add additional metadata validation where in a rename will fail to update the metadata table.  Repair using best guess from the YAML schema?
