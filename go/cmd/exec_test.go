package cmd

import (
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/util"
)

func TestExecDryrun(t *testing.T) {

}

func TestExecRollback(t *testing.T) {

}

func TestExecRollbackInvalid(t *testing.T) {

}

func TestExecPTODisabled(t *testing.T) {

}

func TestExecAllowDestructive(t *testing.T) {

}

func TestExecAllowDestructiveFail(t *testing.T) {

}

func TestExec(t *testing.T) {
	testName := "TestExec"

	util.LogAlert(testName)
	var err error
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	util.SetConfigTesting()

	////////////////////////////////////////////////////////
	// Configure testing data
	//

	// Git requests to pull back state of current checkout

	// GitVersionTime
	gitMySQLTime := "2016-07-12 12:04:05"

	// GitVersionDetails
	gitDetails := `commit abc123
    Author: Scott Porter <sporter@ea.com>
    Date:   Tue Jul 12 22:04:05 2016 +1000

    An example git commit for unit testing`

	// Setup table data
	testConfig := test.GetTestConfig()
	dogsAddTbl := GetTableAddressDogs()

	// Configuring the expected MDID for the new Column
	colMd := dogsAddTbl.Columns[1].Metadata
	colMd.MDID = 4

	// Migration Configuration - use default, standard migration
	dryrun := false
	rollback := false
	PTODisabled := true
	allowDestructive := false

	// Migration id
	mid := int64(1)

	step := migration.Step{
		SID:      1,
		MID:      1,
		Op:       table.Add,
		MDID:     1,
		Name:     "address",
		Forward:  "ALTER TABLE `unittestproject_dogs` COLUMN `address` varchar(128) NOT NULL;",
		Backward: "ALTER TABLE `unittestproject_dogs` DROP COLUMN `address`;",
		Output:   "",
		Status:   migration.Approved,
	}

	m := migration.Migration{
		MID:                1,
		DB:                 1,
		Project:            testConfig.Project.Name,
		Version:            testConfig.Project.Schema.Version,
		VersionTimestamp:   gitMySQLTime,
		VersionDescription: gitDetails,
		Status:             migration.Approved,
		Timestamp:          mysql.GetTimeNow(),
		Steps: []migration.Step{
			step,
		},
		Sandbox: true,
	}

	//
	////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////
	// Configure MySQL access for the management and project DBs
	//

	// Configure the test databases
	// Setup the mock project database
	projectDB, err = test.CreateProjectDB(testName, t)

	if err == nil {
		// Connect to Project DB
		exec.SetProjectDB(projectDB.Db)
	} else {
		t.Errorf("%s failed to setup the Project DB with error: %v", testName, err)
		return
	}

	// Configure the Mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	if err == nil {
		exec.Setup(mgmtDB.Db, 1, testConfig.Project.DB.ConnectString())
		migration.Setup(mgmtDB.Db, 1)
		metadata.Setup(mgmtDB.Db, 1)
	} else {
		t.Errorf("%s failed to setup the Management DB with error: %v", testName, err)
		return
	}

	// Load the requested migration
	mgmtDB.MigrationGet(
		1,
		m.ToDBRow(),
		false,
	)

	// Which will also load it's associated Migration Step
	mgmtDB.MigrationStepGet(
		1,
		step.ToDBRow(),
		false,
	)

	// Get the latest Migration
	mgmtDB.MigrationGetLatest(
		m.ToDBRow(),
		false,
	)

	// Check for running migrations - InProgressID
	mgmtDB.MigrationGetStatus(
		migration.InProgress,
		[]test.DBRow{
			{},
		},
		true,
	)

	// Set this migration to running

	// Update Migration state to InProgress
	mgmtDB.Mock.ExpectExec("update `migration`").WithArgs(
		m.DB,
		testConfig.Project.Name,
		testConfig.Project.Schema.Version,
		m.VersionTimestamp,
		m.VersionDescription,
		migration.InProgress,
		m.MID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update the Migration Step along with the Migration
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		table.Add,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		step.Output,
		migration.Approved,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Load Metadata for the Migration Step operation
	mgmtDB.MetadataGet(
		1,
		colMd.ToDBRow(),
		false,
	)

	// Set Step to InProgress
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		table.Add,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		step.Output,
		migration.InProgress,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	////////////////////////////////////////////////////////////
	// Expect the migration to be executed HERE

	query := test.DBQueryMock{
		Type:   test.ExecCmd,
		Result: sqlmock.NewResult(1, 1),
	}

	query.FormatQuery(step.Forward)
	projectDB.ExpectExec(query)

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Update the Management DB with the result of the migration

	// Set Migration Step to Complete
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		table.Add,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		"Row(s) Affected: 1",
		migration.Complete,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update Metadata
	mgmtDB.MetadataGet(
		1,
		colMd.ToDBRow(),
		false,
	)

	// Update Metadata with completed
	mgmtDB.Mock.ExpectExec("update `metadata`").WithArgs(
		colMd.DB,
		colMd.PropertyID,
		colMd.ParentID,
		colMd.Type,
		colMd.Name,
		true,
		colMd.MDID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update Migration with completed
	mgmtDB.Mock.ExpectExec("update `migration`").WithArgs(
		m.DB,
		testConfig.Project.Name,
		testConfig.Project.Schema.Version,
		m.VersionTimestamp,
		m.VersionDescription,
		migration.Complete,
		m.MID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update the MigrationStep with completed
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		table.Add,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		"Row(s) Affected: 1",
		migration.Complete,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = exec.Exec(exec.Options{
		MID:              mid,
		Dryrun:           dryrun,
		Rollback:         rollback,
		PTODisabled:      PTODisabled,
		AllowDestructive: allowDestructive,
	})

	if err != nil {
		t.Errorf("%s failed with error: %v", testName, err)
		return
	}
}
