CREATE TABLE `admins` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `open_id` varchar(64) DEFAULT NULL COMMENT '微信OpenID',
  `name` varchar(64) DEFAULT NULL,
  `account` varchar(64) NOT NULL,
  `password` varchar(64),
  `permission` varchar(20) NOT NULL COMMENT '权限级别(super最高权限,manager负责人权限,internal内部权限,external外部权限)',
  `point_name` varchar(64) DEFAULT NULL COMMENT '负责点位id',
  `campus` varchar(64) NOT NULL COMMENT '校区(zh朝晖,pf屏峰,mgs莫干山)',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_admins_wx` (`open_id`),
  KEY `idx_admins_point` (`point_name`)
);