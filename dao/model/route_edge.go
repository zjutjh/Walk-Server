package model

import (
	"time"
)

const TableNameRouteEdge = "route_edges"

// RouteEdge mapped from table <route_edges>
type RouteEdge struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	FrontPointID *int64    `gorm:"column:front_point_id;comment:前一个点位ID" json:"front_point_id"`
	PointID      *int64    `gorm:"column:point_id;comment:当前点位ID" json:"point_id"`
	SeqOrder     *int64    `gorm:"column:seq_order;comment:顺序" json:"seq_order"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName RouteEdge's table name
func (*RouteEdge) TableName() string {
	return TableNameRouteEdge
}
