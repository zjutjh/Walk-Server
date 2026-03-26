package main

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/zjutjh/mygo/foundation/command"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gen"
	"gorm.io/gorm"

	"app/register"
)

var tables = []string{
	"user",
	"people",
	"teams",
	"points",
	"routes",
	"route_edges",
	"admins",
	"checkins",
	"wrong_route_records",
}

func main() {
	command.Execute(
		register.Boot,
		func(c *cobra.Command) {},
		func(cmd *cobra.Command, args []string) error { return nil },
	)

	g := gen.NewGenerator(gen.Config{
		OutPath: "./dao/query",
		Mode:    gen.WithDefaultQuery | gen.WithQueryInterface,
	})
	g.UseDB(ndb.Pick())

	m := map[string]func(columnType gorm.ColumnType) (dataType string){
		"tinyint": func(columnType gorm.ColumnType) (dataType string) {
			columnTypeName, ok := columnType.ColumnType()
			if ok && strings.Contains(strings.ToLower(columnTypeName), "unsigned") {
				return "uint8"
			}
			return "int8"
		},
	}
	g.WithDataTypeMap(m)

	for _, table := range tables {
		tableName := g.GenerateModel(table)
		g.ApplyBasic(tableName)
	}

	g.Execute()
}
