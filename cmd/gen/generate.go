package main

import (
	"github.com/spf13/cobra"
	"github.com/zjutjh/mygo/foundation/command"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gen"
	"gorm.io/gen/field"
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
			return "int8"
		},
	}
	g.WithDataTypeMap(m)

	// 有软删除的表
	softDeleteTables := map[string]bool{
		"user": true,
	}

	for _, table := range tables {
		var opts []gen.ModelOpt
		if softDeleteTables[table] {
			opts = append(opts,
				gen.FieldType("deleted_at", "soft_delete.DeletedAt"),
				gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
					return tag.Set("softDelete", "milli")
				}),
				gen.FieldJSONTag("deleted_at", "-"),
			)
		}
		tableName := g.GenerateModel(table, opts...)
		g.ApplyBasic(tableName)
	}

	g.Execute()
}
