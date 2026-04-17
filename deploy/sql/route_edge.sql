CREATE TABLE `route_edges` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `prev_point_name` varchar(64) DEFAULT NULL COMMENT '前一个点位ID',
  `point_name` varchar(64) DEFAULT NULL COMMENT '当前点位ID',
  `route_name` varchar(64) DEFAULT NULL COMMENT '点位所属路线名称', 
  `seq_order` tinyint DEFAULT '0' COMMENT '点位在路线中的顺序',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_route_edges_points` (`prev_point_name`, `point_name`),
  KEY `idx_route_edges_route_point_seq` (`route_name`, `point_name`, `seq_order`)
);