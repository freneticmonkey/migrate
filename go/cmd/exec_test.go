package cmd

import (
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/freneticmonkey/migrate/go/exec"
	"github.com/freneticmonkey/migrate/go/metadata"
	"github.com/freneticmonkey/migrate/go/migration"
	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/table"
	"github.com/freneticmonkey/migrate/go/test"
	"github.com/freneticmonkey/migrate/go/testdata"
	"github.com/freneticmonkey/migrate/go/util"
)

func TestExecDryrun(t *testing.T) {
	testName := "TestExecDryrun"

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
	dogsAddTbl := testdata.GetTableAddressDogs()

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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Verify that the Migration can run

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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Setup for the actual migration

	// Load Metadata for the Migration Step operation
	mgmtDB.MetadataGet(
		1,
		colMd.ToDBRow(),
		false,
	)

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	//

	// The migration shouldn't be run here.
	// Expect NOTHING to be executed HERE

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Update the Management DB with the result of the migration
	// in this case success.

	// NOTHING SHOULD BE UPDATED!

	dryrun = true

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

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)
}

func TestExecRollback(t *testing.T) {
	testName := "TestExecRollback"
	util.LogAlert(testName)

	var err error
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	util.SetConfigTesting()

	////////////////////////////////////////////////////////
	// Configure testing data
	//

	// Build times for testing

	// 1st Aug 2016
	// migrationTimeLatest := mysql.FormatTime(time.Date(2016, 8, 1, 0, 0, 0, 0, time.UTC))
	//
	// gitDetailsLatest := `commit abc123
	// Author: Scott Porter <sporter@ea.com>
	// Date:   Tue Aug 1 00:00:00 2016 +1000
	//
	// An example latest commit dropping the test table`

	// 20th June 2016
	migrationTimeOlder := mysql.FormatTime(time.Date(2016, 7, 20, 0, 0, 0, 0, time.UTC))

	// Git requests to pull back state of current checkout

	// GitVersionDetails
	gitDetailsOlder := `commit abc123
    Author: Scott Porter <sporter@ea.com>
    Date:   Tue Jul 20 00:00:00 2016 +1000

    An example latest commit adding a column to the test table`

	// Setup table data
	testConfig := test.GetTestConfig()
	dogsAddTbl := testdata.GetTableAddressDogs()

	// Configuring the expected MDID for the new Column
	colMd := dogsAddTbl.Columns[1].Metadata
	colMd.MDID = 4

	// Migration Configuration - use default, standard migration
	dryrun := false
	rollback := true
	PTODisabled := true
	allowDestructive := false

	// Define the older Migration

	// Migration id
	olderMID := int64(1)

	olderStep := migration.Step{
		SID:      1,
		MID:      olderMID,
		Op:       table.Add,
		MDID:     1,
		Name:     "address",
		Forward:  "ALTER TABLE `unittestproject_dogs` COLUMN `address` varchar(128) NOT NULL;",
		Backward: "ALTER TABLE `unittestproject_dogs` DROP COLUMN `address`;",
		Output:   "",
		Status:   migration.Approved,
	}

	olderMig := migration.Migration{
		MID:                1,
		DB:                 1,
		Project:            testConfig.Project.Name,
		Version:            testConfig.Project.Schema.Version,
		VersionTimestamp:   migrationTimeOlder,
		VersionDescription: gitDetailsOlder,
		Status:             migration.Approved,
		Timestamp:          migrationTimeOlder,
		Steps: []migration.Step{
			olderStep,
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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Verify that the Migration can run

	// Load the requested migration
	mgmtDB.MigrationGet(
		1,
		olderMig.ToDBRow(),
		false,
	)

	// Which will also load it's associated Migration Step
	mgmtDB.MigrationStepGet(
		1,
		olderStep.ToDBRow(),
		false,
	)

	// Get the latest Migration
	// mgmtDB.MigrationGetLatest(
	// 	latestMig.ToDBRow(),
	// 	false,
	// )

	// Check for running migrations - InProgressID
	mgmtDB.MigrationGetStatus(
		migration.InProgress,
		[]test.DBRow{
			{},
		},
		true,
	)

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Setup for the actual migration

	// Set this migration to running

	// Update Migration state to InProgress
	mgmtDB.Mock.ExpectExec("update `migration`").WithArgs(
		olderMig.DB,
		testConfig.Project.Name,
		testConfig.Project.Schema.Version,
		olderMig.VersionTimestamp,
		olderMig.VersionDescription,
		migration.InProgress,
		olderMig.MID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update the Migration Step along with the Migration - Not effectively doing anything here
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		olderStep.MID,
		olderStep.Op,
		olderStep.MDID,
		olderStep.Name,
		olderStep.Forward,
		olderStep.Backward,
		olderStep.Output,
		migration.Approved,
		olderStep.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Load Metadata for the Migration Step operation
	mgmtDB.MetadataGet(
		1,
		colMd.ToDBRow(),
		false,
	)

	// Now set the step to InProgress

	// Set Step to InProgress
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		olderStep.MID,
		olderStep.Op,
		olderStep.MDID,
		olderStep.Name,
		olderStep.Forward,
		olderStep.Backward,
		olderStep.Output,
		migration.InProgress,
		olderStep.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Expect the migration to be executed HERE

	query := test.DBQueryMock{
		Type:   test.ExecCmd,
		Result: sqlmock.NewResult(1, 1),
	}

	query.FormatQuery(olderStep.Backward)
	projectDB.ExpectExec(query)

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Update the Management DB with the result of the migration
	// in this case success.

	// Set Migration Step to Rollback
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		olderStep.MID,
		olderStep.Op,
		olderStep.MDID,
		olderStep.Name,
		olderStep.Forward,
		olderStep.Backward,
		"Row(s) Affected: 1",
		migration.Rollback,
		olderStep.SID,
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
		false,
		colMd.MDID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update Migration with Rollback
	mgmtDB.Mock.ExpectExec("update `migration`").WithArgs(
		olderMig.DB,
		testConfig.Project.Name,
		testConfig.Project.Schema.Version,
		olderMig.VersionTimestamp,
		olderMig.VersionDescription,
		migration.Rollback,
		olderMig.MID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update the MigrationStep with completed
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		olderStep.MID,
		olderStep.Op,
		olderStep.MDID,
		olderStep.Name,
		olderStep.Forward,
		olderStep.Backward,
		"Row(s) Affected: 1",
		migration.Rollback,
		olderStep.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = exec.Exec(exec.Options{
		MID:              olderMID,
		Dryrun:           dryrun,
		Rollback:         rollback,
		PTODisabled:      PTODisabled,
		AllowDestructive: allowDestructive,
	})

	if err != nil {
		t.Errorf("%s failed with error: %v", testName, err)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)
}

func TestExecFailInvalidRollback(t *testing.T) {
	testName := "TestExecRollbackInvalid"
	util.LogAlert(testName)

	var err error
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	util.SetConfigTesting()

	////////////////////////////////////////////////////////
	// Configure testing data
	//

	// Build times for testing

	// 1st Aug 2016
	migrationTimeLatest := mysql.FormatTime(time.Date(2016, 8, 1, 0, 0, 0, 0, time.UTC))

	gitDetailsLatest := `commit abc123
	Author: Scott Porter <sporter@ea.com>
	Date:   Tue Aug 1 00:00:00 2016 +1000

	An example latest commit dropping the test table`

	// 20th June 2016
	migrationTimeOlder := mysql.FormatTime(time.Date(2016, 7, 20, 0, 0, 0, 0, time.UTC))

	// Git requests to pull back state of current checkout

	// GitVersionDetails
	gitDetailsOlder := `commit abc123
    Author: Scott Porter <sporter@ea.com>
    Date:   Tue Jul 20 00:00:00 2016 +1000

    An example latest commit adding a column to the test table`

	// Setup table data
	testConfig := test.GetTestConfig()
	dogsAddTbl := testdata.GetTableAddressDogs()

	// Configuring the expected MDID for the new Column
	colMd := dogsAddTbl.Columns[1].Metadata
	colMd.MDID = 4

	// Migration Configuration - use default, standard migration
	dryrun := false
	rollback := false
	PTODisabled := true
	allowDestructive := false

	// Define the older Migration

	// Migration id
	olderMID := int64(1)

	olderStep := migration.Step{
		SID:      1,
		MID:      olderMID,
		Op:       table.Add,
		MDID:     1,
		Name:     "address",
		Forward:  "ALTER TABLE `unittestproject_dogs` COLUMN `address` varchar(128) NOT NULL;",
		Backward: "ALTER TABLE `unittestproject_dogs` DROP COLUMN `address`;",
		Output:   "",
		Status:   migration.Approved,
	}

	olderMig := migration.Migration{
		MID:                1,
		DB:                 1,
		Project:            testConfig.Project.Name,
		Version:            testConfig.Project.Schema.Version,
		VersionTimestamp:   migrationTimeOlder,
		VersionDescription: gitDetailsOlder,
		Status:             migration.Approved,
		Timestamp:          migrationTimeOlder,
		Steps: []migration.Step{
			olderStep,
		},
		Sandbox: true,
	}

	// Define the Latest Migration

	// Migration id
	latestMID := int64(1)

	latestStep := migration.Step{
		SID:      2,
		MID:      latestMID,
		Op:       table.Del,
		MDID:     1,
		Name:     "address",
		Forward:  "DROP TABLE `unittestproject_dogs`;",
		Backward: "CREATE TABLE `unittestproject_dogs` BLAH BLAH BLAH;",
		Output:   "",
		Status:   migration.Unapproved,
	}

	latestMig := migration.Migration{
		MID:                latestMID,
		DB:                 1,
		Project:            testConfig.Project.Name,
		Version:            testConfig.Project.Schema.Version,
		VersionTimestamp:   migrationTimeLatest,
		VersionDescription: gitDetailsLatest,
		Status:             migration.Unapproved,
		Timestamp:          migrationTimeLatest,
		Steps: []migration.Step{
			latestStep,
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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Verify that the Migration can run

	// Load the requested migration
	mgmtDB.MigrationGet(
		1,
		olderMig.ToDBRow(),
		false,
	)

	// Which will also load it's associated Migration Step
	mgmtDB.MigrationStepGet(
		1,
		olderStep.ToDBRow(),
		false,
	)

	// Get the latest Migration
	mgmtDB.MigrationGetLatest(
		latestMig.ToDBRow(),
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

	//
	////////////////////////////////////////////////////////////

	err = exec.Exec(exec.Options{
		MID:              olderMID,
		Dryrun:           dryrun,
		Rollback:         rollback,
		PTODisabled:      PTODisabled,
		AllowDestructive: allowDestructive,
	})

	if err == nil {
		t.Errorf("%s Did NOT fail and it should have", testName)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)
}

func TestExecRollbackDryrun(t *testing.T) {
	testName := "TestExecRollbackDryrun"
	util.LogAlert(testName)

	var err error
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	util.SetConfigTesting()

	////////////////////////////////////////////////////////
	// Configure testing data
	//

	// Build times for testing

	// 20th June 2016
	migrationTimeOlder := mysql.FormatTime(time.Date(2016, 7, 20, 0, 0, 0, 0, time.UTC))

	// Git requests to pull back state of current checkout

	// GitVersionDetails
	gitDetailsOlder := `commit abc123
    Author: Scott Porter <sporter@ea.com>
    Date:   Tue Jul 20 00:00:00 2016 +1000

    An example latest commit adding a column to the test table`

	// Setup table data
	testConfig := test.GetTestConfig()
	dogsAddTbl := testdata.GetTableAddressDogs()

	// Configuring the expected MDID for the new Column
	colMd := dogsAddTbl.Columns[1].Metadata
	colMd.MDID = 4

	// Migration Configuration - use default, standard migration
	dryrun := true
	rollback := true
	PTODisabled := true
	allowDestructive := false

	// Define the older Migration

	// Migration id
	olderMID := int64(1)

	olderStep := migration.Step{
		SID:      1,
		MID:      olderMID,
		Op:       table.Add,
		MDID:     1,
		Name:     "address",
		Forward:  "ALTER TABLE `unittestproject_dogs` COLUMN `address` varchar(128) NOT NULL;",
		Backward: "ALTER TABLE `unittestproject_dogs` DROP COLUMN `address`;",
		Output:   "",
		Status:   migration.Approved,
	}

	olderMig := migration.Migration{
		MID:                1,
		DB:                 1,
		Project:            testConfig.Project.Name,
		Version:            testConfig.Project.Schema.Version,
		VersionTimestamp:   migrationTimeOlder,
		VersionDescription: gitDetailsOlder,
		Status:             migration.Approved,
		Timestamp:          migrationTimeOlder,
		Steps: []migration.Step{
			olderStep,
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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Verify that the Migration can run

	// Load the requested migration
	mgmtDB.MigrationGet(
		1,
		olderMig.ToDBRow(),
		false,
	)

	// Which will also load it's associated Migration Step
	mgmtDB.MigrationStepGet(
		1,
		olderStep.ToDBRow(),
		false,
	)

	// Get the latest Migration
	// mgmtDB.MigrationGetLatest(
	// 	latestMig.ToDBRow(),
	// 	false,
	// )

	// Check for running migrations - InProgressID
	mgmtDB.MigrationGetStatus(
		migration.InProgress,
		[]test.DBRow{
			{},
		},
		true,
	)

	//
	////////////////////////////////////////////////////////////

	// Load Metadata for the Migration Step operation
	mgmtDB.MetadataGet(
		1,
		colMd.ToDBRow(),
		false,
	)

	err = exec.Exec(exec.Options{
		MID:              olderMID,
		Dryrun:           dryrun,
		Rollback:         rollback,
		PTODisabled:      PTODisabled,
		AllowDestructive: allowDestructive,
	})

	if err != nil {
		t.Errorf("%s failed with error: %v", testName, err)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)
}

func TestExecPTODisabled(t *testing.T) {

}

func TestExecAllowDestructive(t *testing.T) {
	testName := "TestExecAllowDestructive"

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
	dogsAddTbl := testdata.GetTableAddressDogs()

	// Configuring the expected MDID for the dropped Table
	tableMD := dogsAddTbl.Metadata

	// Migration Configuration - use default, standard migration
	dryrun := false
	rollback := false
	PTODisabled := true
	allowDestructive := true
	force := true // Need to force the application now, otherwise Exec will prompt for approval

	// Migration id
	mid := int64(1)

	step := migration.Step{
		SID:      1,
		MID:      1,
		Op:       table.Del,
		MDID:     1,
		Name:     "unittestproject_dogs",
		Forward:  "DROP TABLE `unittestproject_dogs`;",
		Backward: "CREATE TABLE `unittestproject_dogs` BLAH BLAH BLAH",
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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Verify that the Migration can run

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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Setup for the actual migration

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

	// Update the Migration Step along with the Migration - Not effectively doing anything here
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		step.Op,
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
		tableMD.ToDBRow(),
		false,
	)

	// Now set the step to InProgress

	// Set Step to InProgress
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		step.Op,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		step.Output,
		migration.InProgress,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	//
	////////////////////////////////////////////////////////////

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
	// in this case success.

	// Set Migration Step to Complete
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		step.Op,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		"Row(s) Affected: 1",
		migration.Forced,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update Metadata
	mgmtDB.MetadataGet(
		1,
		tableMD.ToDBRow(),
		false,
	)

	// Mark table in metadata as non-existant
	mgmtDB.Mock.ExpectExec("update `metadata`").WithArgs(
		tableMD.DB,
		tableMD.PropertyID,
		tableMD.ParentID,
		tableMD.Type,
		tableMD.Name,
		false,
		tableMD.MDID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update Migration with completed
	mgmtDB.Mock.ExpectExec("update `migration`").WithArgs(
		m.DB,
		testConfig.Project.Name,
		testConfig.Project.Schema.Version,
		m.VersionTimestamp,
		m.VersionDescription,
		migration.Forced,
		m.MID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Update the MigrationStep with completed
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		step.Op,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		"Row(s) Affected: 1",
		migration.Forced,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = exec.Exec(exec.Options{
		MID:              mid,
		Dryrun:           dryrun,
		Rollback:         rollback,
		PTODisabled:      PTODisabled,
		AllowDestructive: allowDestructive,
		Force:            force,
	})

	if err != nil {
		t.Errorf("%s failed with error: %v", testName, err)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)

}

func TestExecStepApprovedStepUnapproved(t *testing.T) {

}

func TestExecFailAllowDestructive(t *testing.T) {
	testName := "TestExecAllowDestructive"

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
	dogsAddTbl := testdata.GetTableAddressDogs()

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
		Op:       table.Del,
		MDID:     dogsAddTbl.Metadata.MDID,
		Name:     "unittestproject_dogs",
		Forward:  "DROP TABLE `unittestproject_dogs`;",
		Backward: "CREATE TABLE `unittestproject_dogs` BLAH BLAH BLAH",
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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Verify that the Migration can run

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

	//
	////////////////////////////////////////////////////////////

	err = exec.Exec(exec.Options{
		MID:              mid,
		Dryrun:           dryrun,
		Rollback:         rollback,
		PTODisabled:      PTODisabled,
		AllowDestructive: allowDestructive,
	})

	if err == nil {
		t.Errorf("%s SHOULD have FAILED but it did not", testName)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)
}

func TestExecFailMigrationRunning(t *testing.T) {
	testName := "TestExecFailMigrationRunning"

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
	dogsAddTbl := testdata.GetTableAddressDogs()

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
		Op:       table.Del,
		MDID:     dogsAddTbl.Metadata.MDID,
		Name:     "unittestproject_dogs",
		Forward:  "DROP TABLE `unittestproject_dogs`;",
		Backward: "CREATE TABLE `unittestproject_dogs` BLAH BLAH BLAH",
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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Verify that the Migration can run

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
			m.ToDBRow(),
		},
		false,
	)

	//
	////////////////////////////////////////////////////////////

	err = exec.Exec(exec.Options{
		MID:              mid,
		Dryrun:           dryrun,
		Rollback:         rollback,
		PTODisabled:      PTODisabled,
		AllowDestructive: allowDestructive,
	})

	if err == nil {
		t.Errorf("%s SHOULD have FAILED but it did not", testName)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)
}

func TestExecFailDenied(t *testing.T) {
	testName := "TestExecFailDenied"

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
	dogsAddTbl := testdata.GetTableAddressDogs()

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
		Status:   migration.Denied,
	}

	m := migration.Migration{
		MID:                1,
		DB:                 1,
		Project:            testConfig.Project.Name,
		Version:            testConfig.Project.Schema.Version,
		VersionTimestamp:   gitMySQLTime,
		VersionDescription: gitDetails,
		Status:             migration.Denied,
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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Verify that the Migration can run

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

	err = exec.Exec(exec.Options{
		MID:              mid,
		Dryrun:           dryrun,
		Rollback:         rollback,
		PTODisabled:      PTODisabled,
		AllowDestructive: allowDestructive,
	})

	if err == nil {
		t.Errorf("%s SHOULD HAVE FAILED but DID NOT", testName)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)
}

func TestExecFailAlreadyExecute(t *testing.T) {
	testName := "TestExecFailAlreadyExecute"

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
	dogsAddTbl := testdata.GetTableAddressDogs()

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
		Status:   migration.Complete,
	}

	m := migration.Migration{
		MID:                1,
		DB:                 1,
		Project:            testConfig.Project.Name,
		Version:            testConfig.Project.Schema.Version,
		VersionTimestamp:   gitMySQLTime,
		VersionDescription: gitDetails,
		Status:             migration.Complete,
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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Verify that the Migration can run

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

	err = exec.Exec(exec.Options{
		MID:              mid,
		Dryrun:           dryrun,
		Rollback:         rollback,
		PTODisabled:      PTODisabled,
		AllowDestructive: allowDestructive,
	})

	if err == nil {
		t.Errorf("%s SHOULD HAVE FAILED but DID NOT", testName)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)
}

func TestExecFailMissingId(t *testing.T) {
	testName := "TestExecFailMissingId"

	util.LogAlert(testName)
	var err error
	var projectDB test.ProjectDB
	var mgmtDB test.ManagementDB

	util.SetConfigTesting()

	// Migration Configuration - use default, standard migration
	dryrun := false
	rollback := false
	PTODisabled := false
	allowDestructive := false

	////////////////////////////////////////////////////////
	// Configure MySQL access for the management and project DBs
	//

	// Configure the test databases
	// Setup the mock project database
	projectDB, err = test.CreateProjectDB(testName, t)

	if err != nil {
		t.Errorf("%s failed to setup the Project DB with error: %v", testName, err)
		return
	}

	// Configure the Mock Managment DB
	mgmtDB, err = test.CreateManagementDB(testName, t)

	if err != nil {
		t.Errorf("%s failed to setup the Management DB with error: %v", testName, err)
		return
	}

	err = exec.Exec(exec.Options{
		MID:              0, // Invalid MID
		Dryrun:           dryrun,
		Rollback:         rollback,
		PTODisabled:      PTODisabled,
		AllowDestructive: allowDestructive,
	})

	if err == nil {
		t.Errorf("%s didn't fail and it should have", testName)
		return
	}

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)
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
	dogsAddTbl := testdata.GetTableAddressDogs()

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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Verify that the Migration can run

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

	//
	////////////////////////////////////////////////////////////

	////////////////////////////////////////////////////////////
	// Setup for the actual migration

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

	// Update the Migration Step along with the Migration - Not effectively doing anything here
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		step.Op,
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

	// Now set the step to InProgress

	// Set Step to InProgress
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		step.Op,
		step.MDID,
		step.Name,
		step.Forward,
		step.Backward,
		step.Output,
		migration.InProgress,
		step.SID,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	//
	////////////////////////////////////////////////////////////

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
	// in this case success.

	// Set Migration Step to Complete
	mgmtDB.Mock.ExpectExec("update `migration_steps`").WithArgs(
		step.MID,
		step.Op,
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
		step.Op,
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

	projectDB.ExpectionsMet(testName, t)

	mgmtDB.ExpectionsMet(testName, t)
}
