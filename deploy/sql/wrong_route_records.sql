CREATE TABLE `wrong_route_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `team_id` bigint NOT NULL COMMENT '队伍ID',
  `route_name` varchar(64) NOT NULL COMMENT '原正确路线id如pf-half',
  `wrong_route_name` varchar(64) NOT NULL COMMENT '错走的路线id', 
  `admin_id` bigint DEFAULT NULL COMMENT '记录该情况的管理员ID',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_wrong_route_team` (`team_id`),
  KEY `idx_wrong_route_routes` (`route_name`, `wrong_route_name`),
  KEY `idx_wrong_route_time` (`created_at`)
);