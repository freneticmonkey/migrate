package testdata

import (
	"strings"

	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/table"
)

func GetMySQLCreateTableDogs() string {
	var dogsTable = []string{
		"CREATE TABLE `dogs` (",
		"`id` int(11) NOT NULL,",
		" PRIMARY KEY (`id`)",
		") ENGINE=InnoDB DEFAULT CHARSET=latin1;",
	}
	return strings.Join(dogsTable, "\n")
}

func GetCreateTableDogs() string {
	var dogsTable = []string{
		"CREATE TABLE `dogs` (",
		"`id` int(11) NOT NULL,",
		" PRIMARY KEY (`id`)",
		") ENGINE=InnoDB DEFAULT CHARSET=latin1;",
	}
	return strings.Join(dogsTable, "")
}

func GetCreateTableAddressColumnDogs() string {
	var dogsTable = []string{
		"CREATE TABLE `dogs` (",
		"`id` int(11) NOT NULL,",
		"`address` varchar(128) NOT NULL,",
		" PRIMARY KEY (`id`)",
		") ENGINE=InnoDB DEFAULT CHARSET=latin1;",
	}
	return strings.Join(dogsTable, "")
}

func GetYAMLTableDogs() string {
	return `id: dogs
name: dogs
engine: InnoDB
charset: latin1
columns:
- id: id
  name: id
  type: int
  size: [11]
primaryindex:
  id: primarykey
  name: PrimaryKey
  columns:
  - name: id
  isprimary: true
`
}

func GetTableDogs() table.Table {
	return table.Table{
		Name:    "dogs",
		Engine:  "InnoDB",
		CharSet: "latin1",
		Columns: []table.Column{
			{
				Name: "id",
				Type: "int",
				Size: []int{11},
				Metadata: metadata.Metadata{
					MDID:       2,
					DB:         1,
					PropertyID: "id",
					ParentID:   "dogs",
					Name:       "id",
					Type:       "Column",
				},
			},
		},
		PrimaryIndex: table.Index{
			Name:      "PrimaryKey",
			IsPrimary: true,
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
			},
			Metadata: metadata.Metadata{
				MDID:       3,
				DB:         1,
				PropertyID: "primarykey",
				ParentID:   "dogs",
				Name:       "PrimaryKey",
				Type:       "PrimaryKey",
			},
		},
		Metadata: metadata.Metadata{
			MDID:       1,
			DB:         1,
			PropertyID: "dogs",
			Name:       "dogs",
			Type:       "Table",
		},
	}
}

func GetYAMLTableAddressDogs() string {
	return `id: dogs
name: dogs
engine: InnoDB
charset: latin1
columns:
- id: id
  name: id
  type: int
  size: [11]
- id: address
  name: address
  type: varchar
  size: [128]
primaryindex:
  id: primarykey
  name: PrimaryKey
  columns:
  - name: id
  isprimary: true
`
}

func GetTableAddressDogs() table.Table {
	return table.Table{
		Name:    "dogs",
		Engine:  "InnoDB",
		CharSet: "latin1",
		Columns: []table.Column{
			{
				Name: "id",
				Type: "int",
				Size: []int{11},
				Metadata: metadata.Metadata{
					MDID:       2,
					DB:         1,
					PropertyID: "id",
					ParentID:   "dogs",
					Name:       "id",
					Type:       "Column",
				},
			},
			{
				Name: "address",
				Type: "varchar",
				Size: []int{128},
				Metadata: metadata.Metadata{
					// MDID is not defined here as this
					// instance is typically used to test diffing,
					// during which this column needs to be inserted
					// into the DB and as such, the trigger for
					// insertion is MDID < 1
					DB:         1,
					PropertyID: "address",
					ParentID:   "dogs",
					Name:       "address",
					Type:       "Column",
				},
			},
		},
		PrimaryIndex: table.Index{
			Name:      "PrimaryKey",
			IsPrimary: true,
			Columns: []table.IndexColumn{
				{
					Name: "id",
				},
			},
			Metadata: metadata.Metadata{
				MDID:       3,
				DB:         1,
				PropertyID: "primarykey",
				ParentID:   "dogs",
				Name:       "PrimaryKey",
				Type:       "PrimaryKey",
			},
		},
		Metadata: metadata.Metadata{
			MDID:       1,
			DB:         1,
			PropertyID: "dogs",
			Name:       "dogs",
			Type:       "Table",
		},
	}
}
