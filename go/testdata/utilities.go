package testdata

import (
	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/management"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/util"
	"github.com/freneticmonkey/migrate/go/yaml"
)

func Teardown() {
	// Empty Schema
	yaml.Schema = []table.Table{}
	mysql.Schema = []table.Table{}

	// Disassociate Test Databases
	// Connect to Project DB
	exec.SetProjectDB(nil)
	management.SetManagementDB(nil)
	mysql.SetProjectDB(nil)
	exec.Setup(nil, 0, "")
	migration.Setup(nil, 1)
	metadata.Setup(nil, 1)

	// Cleanup util
	util.SetVerbose(false)
	util.ShutdownFileSystem()
}
