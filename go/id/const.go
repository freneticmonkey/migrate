package id

var idConflictTemplate = `
Duplicate ID: >> [%s] found for:
Table: [%s]
Name: [%s]
File: [%s]
-------------
ID already assigned to:
Table:[%s]
Name: [%s]
Type: [%s]
File: [%s]
=============
`

var nameConflictTemplate = `
Duplicate Name: >> [%s] found for:
Table: [%s]
ID: [%s]
File: [%s]
-------------
Name already assigned to:
Table: [%s]
Name: [%s]
Type: [%s]
File: [%s]
=============
`

var missingIDTemplate = `
Missing ID: >> ID: [%s] for:
Name: [%s]
Table: [%s]
Type: [%s]
File: [%s]
`
