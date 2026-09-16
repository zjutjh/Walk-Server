package main

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/zjutjh/mygo/foundation/command"
	"github.com/zjutjh/mygo/ndb"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"app/register"
)

var tables = []string{
	"peoples",
	"teams",
	"points",
	"routes",
	"route_edges",
	"admins",
	"checkins",
	"wrong_route_records",
	"notices",
}

func main() {
	command.Execute(
		register.GenerateBoot,
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
			if ok && strings.Contains(strings.ToLower(columnTypeName), "tinyint(1)") {
				return "bool"
			}
			return "int8"
		},
	}
	g.WithDataTypeMap(m)

	for _, table := range tables {
		opts := []gen.ModelOpt{
			gen.FieldType("deleted_at", "soft_delete.DeletedAt"),
			gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
				return tag.Set("softDelete", "milli")
			}),
			gen.FieldJSONTag("deleted_at", "-"),
		}
		if table == "notices" {
			opts = append(opts,
				gen.FieldType("type", "NoticeType"),
				gen.FieldGenType("type", "String"),
				gen.FieldType("actor_id", "*int64"),
				gen.FieldType("team_id", "*int64"),
				gen.FieldType("read_at", "*time.Time"),
			)
		}
		tableName := g.GenerateModel(table, opts...)
		g.ApplyBasic(tableName)
	}

	g.Execute()
}
