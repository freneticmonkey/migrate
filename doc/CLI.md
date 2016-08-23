
# CLI Design

## global flags
Global Flags are common across all subcommands.

> ### config-file
  Filepath to YAML / JSON Configuration.  If no path is specified the tool defaults to 'config.yml'.

> ### config-url
  URL to YAML / JSON Configuration.  If no URL is specified the tool uses the local 'config.yml' file.

> ### verbose
  Enable all log output

## sandbox
Apply a migration to the database within the sandbox.  Optionally fully recreate the sandbox.  If this is used in production your database is at risk.

### flags

>  ### migrate
   Apply any changes to the local YAML schema to the target project database

>  ### recreate
   Recreate the target database from the YAML Schema in the working folder and insert the metadata into the management database

>  ### dryrun
   Perform a migration dryrun

>  ### force
   Skip the confirmation check before wiping the database and rebuilding the schema

>  ### pull-diff (Optional value) Table Name
   Read the state of the MySQL Target DB and serialise to YAML.  Intended to be used by MySQL power users to store manual schema changes.

## setup
The setup subcommand is used for configuring the migration environment.  The flags to this command determine which environment is being configured.

### flags

>  ### init-management
   Create the management tables in the management database

>  ### init-existing
   Read the target database and generate a YAML schema including PropertyIds

>  ### check-config
   Check the configuration, connectivity to target and management DBs, as well as checking the environment for required tooling.

## diff
Compare the target database to the YAML schema and output a human readable Git style diff.

### flags
If the following flags aren't defined then the contents of the working directory is used

> ### project
  The target project

> ### version
  The target git version

> ### table
  Diff a specific table

## validate
Process the YAML schema and the target database and detail any problems such as missing PropertyIds or invalid YAML schema.  The number of issues found is returned.

### flags
If the following flags aren't defined then the contents of the working directory is used

> ### project
  The target project

> ### version
  The target git version

> ### schema-type
  Specific schema to validate (mysql,yaml).  Defaults to 'both'

## create
This subcommand is used to create a migration and register it with the management database.  Migrations are defined using a project name and git version hash.  Each migration is assigned an identifier by the management database, which is used by the **exec** subcommand to select the migration to apply.

### flags
> ### version
  The target git version

> ### rollback
  Allows for a rollback migration to be created

## exec
Migrations created by the **create** are executed by this subcommand.  Migrations are identified by an id.  The *dryrun* flag ensures that the migration is only tested and not applied to the target database.

### flags
> ### id
  The id of the migration to execute

> ### print
  Display the current state of the migration with ID as recorded in the management DB.

> ### dryrun
  Execute a dryrun of the migration

> ### rollback
  Allows for a rollback migration to be executed

> ### pto-disabled
  Execute the migration without using pt-online-schema-change

> ### allow-destructive
  Specifically allow migrations containing rename or delete actions.

## serve
Starts a REST API Server which provides access to the management database.  Optionally, if the --frontend flag is used, the contents of a subfolder named 'static' will also be served.  The REST API provides endpoints for listing Migrations and Migration Steps, and allows for the status of Migration and Migration Steps to be updated.

## flags
> ### frontend
      If defined, the server will serve the web frontend in addition to the REST API from a subfolder named 'static'

> ### port
      Allows for an alternative port to be used.
