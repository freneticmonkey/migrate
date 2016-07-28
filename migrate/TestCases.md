
# Integration Test Cases

Should start by building tests around standard use cases and then, brain storm how
game teams will want to use this. Add in additional tests for executing things things out of order, invalid or misconfigured setups.

# Unit Test data mockup framework
- [ ] DB Mockup helper
    - [ ] metadata
    - [ ] database
    - [ ] migration
- [ ] Git commands result helper

# Sandbox

- [ ] Refresh database
- [ ] Make and apply a table or column change immediately
- [ ] Define a table with bare minimum properties. No options defined.  Default options test.
- [ ] Execute a dryrun of
    - [ ] a setup
    - [ ] a change
- [ ] Test that force will only work in a database configured as a sandbox

# Setup
- [ ] Setup the management database from an existing database
- [ ] Setup the management database

# Diff
- [ ] Check that the diff command works with changes to:
    - [ ] Tables
    - [ ] Columns
    - [ ] Indexes

# Create
- [ ] Test creating valid forward migration with project and version
- [ ] Test creating valid forward migration with only project testing head version extraction from git
- [ ] Test creating valid backward migration using --rollback
- [ ] Test FAIL: creating old (invalid) forward migration without rollback
- [ ] Test FAIL: no project provided.

# Validate
- [ ] Test valid YAML
- [ ] Test valid MySQL
- [ ] Test malformed YAML
- [ ] Test malformed MySQL

# Exec
- [ ] Test applying:
    - [ ] a valid migration using a dryrun
    - [ ] a valid backward migration using a dryrun
    - [ ] a valid forward migration
    - [ ] a valid backward migration
    - [ ] a destructive migration with --allow-destructive
    - [ ] FAIL: invalid (old) forward migration
    - [ ] FAIL: an already applied migration
    - [ ] FAIL: a migration while another migration is in progress
    - [ ] FAIL: an unknown migration (unknown id)
    - [ ] FAIL: a destructive migration WITHOUT --allow-destructive
    - [ ] FAIL: mid migration resulting in a rename not updating the metadata in the management database.
