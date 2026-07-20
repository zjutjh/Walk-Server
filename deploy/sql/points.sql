CREATE TABLE `points` (
  `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
  `cp_id`      bigint unsigned NOT NULL COMMENT '校区内点位编号(可跨校区重复)',
  `name`       varchar(64)     DEFAULT NULL COMMENT '全局唯一点位名称,拼音首字母,如jls（金莲寺）',
  `is_active`  tinyint(1)      DEFAULT '1' COMMENT '是否启用',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_points_name` (`name`),
  KEY `idx_points_cid` (`cp_id`)
);