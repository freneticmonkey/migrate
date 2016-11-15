# Getting Started

## Setup migrate

Install Go 1.7

Clone and build migrate

```
git clone https://github.com/freneticmonkey/migrate
cd migrate
go build -o examples/migrate ./go
```

## Creating the example project

The example folder contains a sample project with a minimal defined schema that can be used to test migrate features.

The sample project requires Docker so please install it before continuing.

```
cd example
docker-compose up -d
```

This will start two docker containers.  The `management_db` is used to store the schema metadata and migrations for a project.  The `target_db` hosts the project database that will be migrated in the example.


## Setting up migrate

By default Migrate is configured using `config.yml` alongside the migrate executable.  Configuration files can be loaded using either of the `--config-file` or `--config-url` options which will load a configuration file by path or from a URL.

### Config.yml
Configuration within  this file is used to configure the management and project databases, in addition to working paths, and schema generation options.

## Creating the management schema

First the management_db needs a schema to store the management data used by migrate.  

Create the database in the management MySQL instance

`docker exec -t -i example_management_db_1 mysql -ptest -e 'create database management;'`

Then, run migrate to initialise the management schema.

`./migrate setup --management`

## Creating the project schema

Create the project database

`docker exec -t -i example_target_db_1 mysql -ptest -e 'create database test;'`

At this point you can check to see what changes are going to be executed by migrate by running a diff.

`./migrate diff`

Which will output:

```
+++ CREATE TABLE `cats`
(
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `name` varchar(64) NOT NULL,
    `age` int(11) NOT NULL,
    PRIMARY KEY (`name`),
    KEY `idx_id_name` (`id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
+++ CREATE TABLE `dogs`
(
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `name` varchar(64) NOT NULL,
    `age` int(11) NOT NULL,
    `address` varchar(256) NOT NULL,
    PRIMARY KEY (`name`),
    KEY `idx_id_name` (`id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
```
*formatting added for readability. It will be one line per operation normally*

Execute the migration in the sandbox

`./migrate sandbox --migrate`

To view the new tables and their schema

```
docker exec -t -i example_target_db_1 mysql -ptest -e 'show tables;show create table `cats`; show create table `dogs`' test
```

This will show the following output:

```
+----------------+
| Tables_in_test |
+----------------+
| cats           |
| dogs           |
+----------------+
+-------+--------------------------------+
| Table | Create Table                   |
+-------+--------------------------------+
| cats  | CREATE TABLE `cats` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `age` int(11) NOT NULL,
  PRIMARY KEY (`name`),
  KEY `idx_id_name` (`id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1   |
+-------+--------------------------------+
+-------+--------------------------------+
| Table | Create Table                   |
+-------+--------------------------------+
| dogs  | CREATE TABLE `dogs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `age` int(11) NOT NULL,
  `address` varchar(256) NOT NULL,
  PRIMARY KEY (`name`),
  KEY `idx_id_name` (`id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1   |
+-------+--------------------------------+

```

## Altering a table in the sandbox

The sandbox environment is different to other enviroments in that the verification process is not required to apply migration changes.  In order to achieve this there is a separate subcommand `sandbox` which provides additional sandbox specific operations.  For more info see the CLI documentation.

The setup will have created a migration for the initial create table statements, however what we really want to look at is how migrations are created inside the sandbox.

Open working/animals/cats.yml and add the following column to the columns array

```

- name:     address
  type:     int
  size:     [11]
  nullable: No
  id:       address

```

The `diff` operation can be used to view the differences between the local yaml schema at the project database schema.

`./migrate diff`

You will see the following output for the new column.

```
+++ ALTER TABLE `cats` ADD COLUMN `address` int(11) NOT NULL;
```

Apply the change using the same command with which you created the inital schema.

`./migrate sandbox --migrate`

The database schema will now reflect the change.

```
docker exec -t -i example_target_db_1 mysql -ptest -e 'show tables;show create table `cats`; show create table `dogs`' test
```

Result:

```
+----------------+
| Tables_in_test |
+----------------+
| cats           |
| dogs           |
+----------------+
+-------+--------------------------------+
| Table | Create Table                   |
+-------+--------------------------------+
| cats  | CREATE TABLE `cats` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `age` int(11) NOT NULL,
  `address` int(11) NOT NULL,
  PRIMARY KEY (`name`),
  KEY `idx_id_name` (`id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1   |
+-------+--------------------------------+
+-------+--------------------------------+
| Table | Create Table                   |
+-------+--------------------------------+
| dogs  | CREATE TABLE `dogs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `age` int(11) NOT NULL,
  `address` varchar(256) NOT NULL,
  PRIMARY KEY (`name`),
  KEY `idx_id_name` (`id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1   |
+-------+--------------------------------+
```

## Creating a migration

Outside of the sandbox environment migrations cannot be applied without validation.  Migrations are created, viewed, and then approved before they can be applied to the project database.

Migrations are created with the `create` subcommand.  The example below shows the use of additional flags `--gifinfo` which provides a 'gitinfo' file in place of a local repository, and `--no-clone` which disabled git cloning and relies on any schema located within the working directory.

`./migrate create --gitinfo gitinfo.txt --no-clone`

This command will create a migration in the management database.  An example of this is shown below:

```
+-----+----+---------+------------------------------------------+---------------------+--------------------------------------------------+--------+---------------------+
| mid | db | project | version                                  | version_timestamp   | version_description                              | status | timestamp           |
+-----+----+---------+------------------------------------------+---------------------+--------------------------------------------------+--------+---------------------+
|   1 |  1 | animals |                                          | 2016-11-15 12:54:11 | Sandbox Migration                                |      5 | 2016-11-15 10:26:21 |
|   2 |  1 | animals | 2932b57948f65a2e9a97713fe51718a86d6acabc | 2016-11-16 03:55:56 | commit 2932b57948f65a2e9a97713fe51718a86d6acabc
Author: freneticmonkey <scottporter@neuroticstudios.com>
Date:   Tue Nov 16 14:55:56 2016 +1100

    Fixed slight verbose logging for YAML issues.
 |      0 | 2016-11-15 10:30:06 |
+-----+----+---------+------------------------------------------+---------------------+--------------------------------------------------+--------+---------------------+
+-----+------+------+------+-----------+-----------------------------------------------------------------+---------------------------------------------+--------+--------+
| sid | mid  | op   | mdid | name      | forward                                                         | backward                                    | output | status |
+-----+------+------+------+-----------+-----------------------------------------------------------------+---------------------------------------------+--------+--------+
|   1 |    2 |    0 |   16 | something | ALTER TABLE `cats` ADD COLUMN `something` varchar(64) NOT NULL; | ALTER TABLE `cats` DROP COLUMN `something`; |        |      0 |
+-----+------+------+------+-----------+-----------------------------------------------------------------+---------------------------------------------+--------+--------+
```

## Applying a migration

Applying a migration doesn't require the yaml schema, the migration process just applies each of the approved steps of a migration to the project database.  Migrations that have been approved cannot be executed if they are outdated by a newer migration.

Below is an example of a migration using the `--gitinfo` flag as in the previous example to select the migration associated with the appropriate git version.

`./migrate exec --gitinfo gitinfo.txt`

This will produce the following output.

```
TODO:
```

The logs generated by a migration operation are written to the management database into each migration step as they are being applied.

## Rolling back migrations

TODO:

`./migrate `
