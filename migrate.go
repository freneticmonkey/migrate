package main

import (
	"strings"

	"github.com/freneticmonkey/migrate/migrate/config"
	"github.com/freneticmonkey/migrate/migrate/util"
)

var conf config.Config

type Versions []string

func (i *Versions) String() string {
	return strings.Join(*i, ",")
}

func (i *Versions) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// var flags struct {
//   // help bool
//   dryrun   bool
// 	showdiff bool
// 	config   string
//   logfile  string
//
//   git      bool
//
//   versions  Versions
// }
//
// func processFlags() {
//   flags.dryrun    = false
//   flags.showdiff  = false
//   flags.config    = "config/config.yml"
//   flags.logfile   = ""
//
//   flag.BoolVar(&flags.dryrun, "dryrun", false, "Perform a dry run")
//   flag.BoolVar(&flags.showdiff, "diff", false, "Show differences")
//   flag.StringVar(&flags.config, "config", flags.config, "Define config file (YAML)")
//   flag.StringVar(&flags.logfile, "logfile", "", "log file name")
//
//   flag.BoolVar(&flags.git, "git", false, "execute the git functionality")
// 	flag.Var(&flags.versions, "version", "git version to checkout using the format <name>:<version>")
//
//   flag.Parse()
// }
//
// // Read Configuration
// func readConfig() {
//   err := migrate.ReadYamlFile(flags.config, &config)
//   migrate.ErrorCheck(err)
//
//   if len(flags.versions) > 0 {
//     config.Schema.SetVersions(flags.versions)
//   }
//
//   migrate.LogInfo("Configuration Read Success: " + flags.config)
// }
//
// func processLogging() {
//   // Configure Logging
// 	if flags.logfile != "" {
// 		f, err := os.OpenFile(flags.logfile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
// 		migrate.ErrorCheck(err)
//
// 		defer f.Close()
// 		log.SetOutput(f)
// 	}
// }

func main() {
	// processFlags()

	util.LogInfo("Migrations!")

	// processLogging()
	//
	// readConfig()
	//
	// if flags.git {
	//   LogInfo("Running Git functions")
	//   migrate.GitClone(config)
	// }
	//
	// migrate.ReadYamlTables(config.Schema)
	//
	// migrate.ReadDBTables(config.DB)
	//
	// differences := migrate.DiffTables(migrate.YamlTables.Tables, migrate.DBSchema.Tables)
	//
	// migrate.GenerateMySQLAlters(differences)

	//migrate.WriteDBTables(config.Yaml.Path, migrate.DBSchema.Tables)
}
