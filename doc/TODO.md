# TODO

A growing list of features and tasks that need to be completed before this project is ready for use.

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
- [x] DEFAULT values support
- [x] Partial string index support
- [x] ROW_FORMAT support
- [x] Column Collation support
- [x] Table Default Collation support

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
- [ ] Add MySQL validation option for validate cli

# TODO Polish
- [x] Provide feedback when flags are required
- [x] Setup --test flag for testing configuration (DB)
- [x] Add a dependency check flag to setup which checks env for git and pt-online-schema-change
- [x] Migration inspection which displays migrations in cli
- [x] Diff a specific table
- [x] Pull Schema changes back to the YAML (Manual sandbox edits)


# TODO Future / Maybe
- [ ] Add strict mode which hashes the create table statement and stores it for validation against each of the migration_steps.
- [ ] Add additional metadata validation where in a rename will fail to update the metadata table.  Repair using best guess from the YAML schema?
