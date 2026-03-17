CREATE TABLE `checkins` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `admin_id` bigint DEFAULT NULL COMMENT '签到管理员ID',
  `team_id` bigint NOT NULL COMMENT '队伍ID',
  `point_name` varchar(64) NOT NULL COMMENT '签到点位ID',
  `route_name` varchar(64) DEFAULT NULL COMMENT '路线id',
  `time` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '签到时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_checkins_route_point` (`route_name`, `point_name`),
  KEY `idx_checkins_time` (`time`)
);