package yaml

import (
	"github.com/freneticmonkey/migrate/go/config"
	"github.com/freneticmonkey/migrate/go/table"
)

// Schema The parsed from the YAML tables
var Schema table.Tables

var useNamespaces bool

func Setup(conf config.Config) {
	useNamespaces = conf.Options.Namespaces
}
