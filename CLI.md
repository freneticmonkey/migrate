
# CLI Design

## global flags
Global Flags are common across all subcommands.

> ### config-url
  URL to YAML / JSON Configuration.  If no URL is specified the tool uses the local 'config.yml' file.

> ### verbose
  Enable all log output

## setup
The setup subcommand is used for configuring the migration environment.  The flags to this command determine which environment is being configured.

### flags
>  ### sandbox
   Recreate the target database from the YAML Schema in the working folder and insert the metadata into the target database

>  ### init-management
   Create the management tables in the management database

>  ### init-existing
   Read the target database and generate a YAML schema including PropertyIds

## diff
Compare the target database to the YAML schema and output a human readable Git style diff.

### flags
If the following flags aren't defined then the contents of the working directory is used

> ### project
  The target project

> ### version
  The target git version

## validate
Process the YAML schema and the target database and detail any problems such as missing PropertyIds or invalid YAML schema.  The number of issues found is returned.

### flags
If the following flags aren't defined then the contents of the working directory is used

> ### project
  The target project

> ### version
  The target git version

## create
This subcommand is used to create a migration and register it with the management database.  Migrations are defined using a project name and git version hash.  Each migration is assigned an identifier by the management database, which is used by the **exec** subcommand to select the migration to apply.

### flags
> ### project
  The target project

> ### version
  The target git version

> ### rollback
  Allows for a rollback migration to be created

> ### force-sandbox
  Immediately apply the new migration to the target database.  Will only function in the sandbox environment.

## exec
Migrations created by the **create** are executed by this subcommand.  Migrations are identified by an id.  The *dryrun* flag ensures that the migration is only tested and not applied to the target database.

### flags
> ### id
  The id of the migration to execute

> ### dryrun
  Execute a dryrun of the migration

> ### rollback
  Allows for a rollback migration to be executed
