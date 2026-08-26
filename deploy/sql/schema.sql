-- Walk-Server database schema
-- Import with a UTF-8 client, for example:
-- mysql --default-character-set=utf8mb4 -u root -p walk_server < schema.sql

SET NAMES utf8mb4;

CREATE TABLE `admins` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) DEFAULT NULL,
  `account` varchar(64) NOT NULL,
  `password` varchar(64),
  `permission` varchar(20) NOT NULL COMMENT '权限级别(super最高权限,manager负责人权限,internal内部权限,external外部权限)',
  `point_name` varchar(64) DEFAULT NULL COMMENT '负责点位id',
  `campus` varchar(64) NOT NULL COMMENT '校区(zh朝晖,pf屏峰,mgs莫干山)',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_admins_account` (`account`),
  KEY `idx_admins_point` (`point_name`)
);

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

CREATE TABLE `points` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `cp_id` bigint unsigned NOT NULL COMMENT '校区内点位编号(可跨校区重复)',
  `name` varchar(64) DEFAULT NULL COMMENT '全局唯一点位名称,拼音首字母,如jls（金莲寺）',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否启用',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_points_name` (`name`)
);

CREATE TABLE `route_edges` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `prev_point_name` varchar(64) DEFAULT NULL COMMENT '前一个点位ID',
  `point_name` varchar(64) DEFAULT NULL COMMENT '当前点位ID',
  `route_name` varchar(64) DEFAULT NULL COMMENT '点位所属路线名称',
  `seq_order` tinyint DEFAULT '0' COMMENT '点位在路线中的顺序',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_route_edges_point_route` (`point_name`, `route_name`),
  KEY `idx_route_edges_route_prev_point` (`route_name`, `prev_point_name`, `point_name`),
  KEY `idx_route_edges_route_point_seq` (`route_name`, `point_name`, `seq_order`),
  KEY `idx_route_edges_route_seq` (`route_name`, `seq_order`)
);

CREATE TABLE `teams` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '队伍ID',
  `name` varchar(64) NOT NULL COMMENT '队伍名称',
  `num` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '团队人数',
  `password` varchar(64) NOT NULL COMMENT '团队加入密码',
  `slogan` varchar(128) DEFAULT NULL COMMENT '团队标语',
  `allow_match` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否允许随机匹配',
  `captain` bigint unsigned NOT NULL COMMENT '队长用户ID',
  `submit` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否已提交报名',
  `route_name` varchar(64) DEFAULT NULL COMMENT '团队所属路线',
  `latest_point_name` varchar(64) DEFAULT NULL COMMENT '最新经过点位ID',
  `status` varchar(64) NOT NULL COMMENT '活动状态(notStart未出发，inProgress进行中，completed已完成，withdrawn已下撤)',
  `is_wrong_route` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否走错',
  `is_reunite` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否重组',
  `code` varchar(128) DEFAULT NULL COMMENT '签到二维码绑定码',
  `time` datetime(3) DEFAULT NULL COMMENT '队伍状态更新时间',
  `is_lost` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否失联',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_teams_code` (`code`),
  KEY `idx_teams_submit_route` (`submit`, `route_name`),
  KEY `idx_teams_submit_wrong_route_name` (`submit`, `is_wrong_route`, `route_name`),
  KEY `idx_teams_route_point` (`route_name`, `latest_point_name`),
  KEY `idx_teams_route_match_num` (`route_name`, `allow_match`, `num`)
);

CREATE TABLE `peoples` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `password` varchar(60) NOT NULL DEFAULT '' COMMENT 'bcrypt密码哈希，空值表示预导入校友尚未注册',
  `name` varchar(128) NOT NULL COMMENT '姓名',
  `gender` tinyint NOT NULL DEFAULT '0' COMMENT '性别(0未知,1男,2女)',
  `stu_id` varchar(32) DEFAULT NULL COMMENT '学号',
  `identity` varchar(128) NOT NULL COMMENT '身份证号AES密文',
  `role` varchar(64) NOT NULL DEFAULT 'unbind' COMMENT '队伍中身份(unbind未绑定,menber成员,captain队长)',
  `qq` varchar(20) DEFAULT NULL COMMENT 'QQ号',
  `wechat` varchar(64) DEFAULT NULL COMMENT '微信号',
  `tel` varchar(20) NOT NULL COMMENT '联系电话',
  `created_op` tinyint unsigned NOT NULL DEFAULT '3' COMMENT '剩余创建团队次数',
  `join_op` tinyint unsigned NOT NULL DEFAULT '5' COMMENT '剩余加入团队次数',
  `team_id` bigint DEFAULT '-1' COMMENT '所属团队ID',
  `type` varchar(10) NOT NULL COMMENT '人员类型(alumni校友，student学生，teacher教职工)',
  `walk_status` varchar(64) NOT NULL COMMENT '活动状态(未开始,待出发,已放弃,进行中,已下撤,已完成)',
  `is_violated` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否违规',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_people_identity` (`identity`),
  UNIQUE KEY `uni_people_tel` (`tel`),
  UNIQUE KEY `uni_people_stu_id` (`stu_id`),
  KEY `idx_people_team_walk_status` (`team_id`, `walk_status`),
  KEY `idx_people_walk_status` (`walk_status`)
);

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
  KEY `idx_checkins_team_point` (`team_id`, `point_name`),
  KEY `idx_checkins_team_time` (`team_id`, `time`)
);

CREATE TABLE `wrong_route_records` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `team_id` bigint NOT NULL COMMENT '队伍ID',
  `route_name` varchar(64) NOT NULL COMMENT '原正确路线id如pf-half',
  `wrong_route_name` varchar(64) NOT NULL COMMENT '错走的路线id',
  `admin_id` bigint DEFAULT NULL COMMENT '记录该情况的管理员ID',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_wrong_route_team_time` (`team_id`, `created_at`)
);
