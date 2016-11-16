
# REST API

### Standard Return Format
All responses from the REST API will return in a standard format which allows for errors and response to be easily detected.

    {
      "result": {
        <response JSON>
      },
      "error": {
        "error": <string error>,
        "detail": <additional error information>
      }
    }

## Endpoints

### /api/migration/
Interact with migrations in the Management DB

#### /api/migration/{id}
Get the Migration with ID {id}

#### /api/migration/list/
Get the first 10 Migrations

#### /api/migration/list/{start}
Get the first 10 Migrations from id {start}

#### /api/migration/list/{start}/{count}
Get all Migrations with IDs between {start} and {start} + {count}

#### /api/status/edit/
Update Migration and Step Status can be updated with the following POST structure

    {
        "vetted_by" : "<name of vetter>",
        "migrations": [
            {
                "mid" : <mid>,
                "status" : <new_status>
            },
            ...
        ],
        "steps": [
            {
                "sid" : <mid>,
                "status": <new_status>
            },
            ...
        ]
    }

### /api/database/
Interact with database(s) defined within the Management DB

#### /api/database/{id}
Get the Database with ID {id}

### /api/table/
This endpoint provides the ability to create/edit/diff/delete(drop) tables.
The JSON structure for tables can be seen below.

    {
        "id": "person",
    	"name": "person",
    	"engine": "InnoDB",
    	"autoinc": 0,
    	"charset": "latin1",
        "columns" : [
            {
                "id" : "id",
                "type" : "int",
                "size" : [11],
                "autoinc" : true,
                "nullable" : false
            },
            {
                "id" : "name",
                "type" : "varchar",
                "size" : [64],
                "nullable" : false
            },
            {
                "id" : "address",
                "type" : "varchar",
                "size" : [128],
                "nullable" : false
            }
            ...
        ],
        "primaryindex": {
            "id" : "primarykey",
            "name" : "primarykey",
            "columns" : [
                {
                    "name" : "id"
                }
            ]
        },
        "secondaryindexes" : [
            {
                "id" : "idx_name",
                "name" : "idx_name",
                "columns" : [
                    {
                        "name" : "name",
                        "length" : 12
                    }
                ]
            }
            ...
        ]
    }

#### /api/table/list/
Get the first 10 tables

#### /api/table/list/{start}
Get the first 10 tables from id {start}

#### /api/table/list/{start}/{count}
Get all Tables with with pagination offsets {start} and {start} + {count}

#### /api/table/create/
Create a new Table given the Table definition passed to this endpoint generating
a new YAML file.

#### /api/table/{id}/edit/
Edit a Table with ID {id} generating a new YAML file with the changes

#### /api/table/{id}/delete/
Delete the YAML file for the table with ID {id}

### /api/sandbox/
This endpoint contains utilities for manipulating the schema in the sandbox

#### /api/sandbox/diff/{id}
Generate the SQL ALTER TABLE statements for the Table with {id} or for all Tables if {id} is not defined

#### /api/sandbox/migrate/
Apply differences in the YAML schema state to the target DB

#### /api/sandbox/recreate/
Erase sandbox DB and recreate it from the YAML schema

#### /api/sandbox/pull-diff/
Serialise manual alterations of the MySQL Schema to YAML files

### /api/health/
Health check using the setup --check-config functionality
