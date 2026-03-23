package repo

import (
	"github.com/zjutjh/mygo/ndb"

	"app/dao/query"
)

func newQuery() *query.Query {
	return query.Use(ndb.Pick())
}
