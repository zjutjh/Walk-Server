CREATE TABLE `routes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL COMMENT '路线代码，如pf-half, pf-full等',
  `point_name` varchar(64) NOT NULL COMMENT '路线名称，如屏峰半程',
  `campus` varchar(64) NOT NULL COMMENT '校区(zh朝晖,pf屏峰,mgs莫干山)',
  `is_active` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_routes_name` (`name`),
  KEY `idx_routes_campus_active` (`campus`, `is_active`)
);