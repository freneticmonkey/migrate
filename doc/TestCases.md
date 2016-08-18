
# Integration Test Cases

Should start by building tests around standard use cases and then, brain storm how
game teams will want to use this. Add in additional tests for executing things things out of order, invalid or misconfigured setups.

# Unit Test data mockup framework
- [x] DB Mockup helper
    - [x] metadata
    - [ ] database
    - [x] migration
- [x] Git commands result helper

# Sandbox

- [x] Refresh database
- [ ] Make and apply a table or column change immediately
- [x] Define a table with bare minimum properties. No options defined.  Default options test.
- [ ] Execute a dryrun of
    - [ ] a setup
    - [ ] a change
- [ ] Test that force will only work in a database configured as a sandbox

# Setup
- [x] Setup the management database from an existing database
- [x] Setup the management database

# Diff
- [x] Check that the diff command works with changes to:
    - [x] Tables
    - [x] Columns
    - [x] Indexes

# Create
- [x] Test creating valid forward migration with project and version
- [ ] Test creating valid forward migration with only project testing head version extraction from git
- [ ] Test creating valid backward migration using --rollback
- [ ] Test FAIL: creating old (invalid) forward migration without --rollback
- [x] Test FAIL: no project provided.

# Validate
- [x] Test valid YAML
- [x] Test valid MySQL
- [x] Test malformed YAML
- [x] Test malformed MySQL

# Exec
- [ ] Test applying:
    - [x] a valid migration using a dryrun
    - [x] a valid backward migration using a dryrun
    - [x] a valid forward migration
    - [x] a valid backward migration
    - [x] a destructive migration with --allow-destructive
    - [x] FAIL: invalid (old) forward migration
    - [x] FAIL: an already applied migration
    - [ ] FAIL: a migration while another migration is in progress
    - [x] FAIL: an unknown migration (unknown id)
    - [x] FAIL: a destructive migration WITHOUT --allow-destructive
    - [ ] FAIL: mid migration fail during a rename which results in the metadata not being updated in the management database.

# YAML
- [ ] Namespacing
    - [ ] Nested folders
