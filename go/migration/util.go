package migration

import (
	"fmt"
	"time"

	"github.com/freneticmonkey/migrate/go/mysql"
	"github.com/freneticmonkey/migrate/go/util"
)

// VersionExists Check if the Git version has already been registered for migration
func VersionExists(hash string) (exists bool, err error) {
	var count int64
	query := fmt.Sprintf("select count(*) from migration WHERE version = \"%s\"", hash)
	count, err = mgmtDb.SelectInt(query)
	if err == nil {
		exists = (count > 0)
	}
	return exists, err
}

// IsLatest Return if the RFC3339 formatted timestamp is newer than the newest migration in the DB
func IsLatest(newTime string) (isLatest bool, err error) {
	var checkTime time.Time
	var latestDBTime time.Time
	var latest Migration

	checkTime, err = time.Parse(mysql.TimeFormat, newTime)
	if !util.ErrorCheckf(err, "Unable to parse time parameter. Time: [%s]", newTime) {
		latest, err = GetLatest()

		if !util.ErrorCheck(err) {
			latestDBTime, err = time.Parse(mysql.TimeFormat, latest.VersionTimestamp)

			if !util.ErrorCheckf(err, "Unable to parse latest time accordig to the DB. DB Time: [%s]", latest.VersionTimestamp) {
				isLatest = checkTime.After(latestDBTime)
			}
		}
	}

	return isLatest, err
}

// HasMigrations There are existing migrations
func HasMigrations() (result bool, err error) {
	var count int64
	count, err = mgmtDb.SelectInt("select count(*) from migration")
	if !util.ErrorCheckf(err, "Unable to check for existing Migrations in the Management DB") {
		result = (count > 0)
	}
	return result, err
}

// GetLatest Return the git timestamp latest Migration from the DB
func GetLatest() (m Migration, err error) {
	var migrations Migration
	err = mgmtDb.SelectOne(&migrations, "select * from migration ORDER BY version_timestamp DESC LIMIT 1")
	util.ErrorCheckf(err, "Unable to get latest Migration from Management DB")
	return m, err
}
