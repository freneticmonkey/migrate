
# REST API

## Standard Return Format
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

### /api/migration/{id}
Get the Migration with ID {id}

### /api/migration/list/{start}/{count}
Get all Migrations with IDs between {start} and {start} + {count}

### /api/status/edit/
Update Migration and Step Status can be updated with the following POST structure

    {
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

### /api/database/{id}
Get the Database with ID {id}
