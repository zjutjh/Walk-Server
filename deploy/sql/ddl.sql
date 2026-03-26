-- =============================================
-- 毅行系统数据库DDL
-- =============================================

-- 1. People（人员基础信息表）
CREATE TABLE `people` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `open_id`     VARCHAR(64)     NOT NULL COMMENT '微信OpenID',
    `name`        VARCHAR(128)    NOT NULL COMMENT '姓名',
    `gender`      TINYINT         NOT NULL COMMENT '性别(1男,2女)',
    `stu_id`      VARCHAR(32)     DEFAULT NULL COMMENT '学号',
    `campus`      TINYINT UNSIGNED NOT NULL COMMENT '校区(1朝晖,2屏峰,3莫干山)',
    `identity`    VARCHAR(18)     NOT NULL COMMENT '身份证号',
    `role`        TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '队伍角色(0未加入,1队员,2队长)',
    `qq`          VARCHAR(20)     DEFAULT NULL COMMENT 'QQ号',
    `wechat`      VARCHAR(64)     DEFAULT NULL COMMENT '微信号',
    `college`     VARCHAR(64)     NOT NULL COMMENT '学院',
    `tel`         VARCHAR(20)     NOT NULL COMMENT '联系电话',
    `created_op`  TINYINT UNSIGNED NOT NULL DEFAULT 3 COMMENT '创建团队剩余次数',
    `join_op`     TINYINT UNSIGNED NOT NULL DEFAULT 5 COMMENT '加入团队剩余次数',
    `team_id`     BIGINT          DEFAULT -1 COMMENT '所属团队ID',
    `type`        TINYINT UNSIGNED NOT NULL COMMENT '人员类型(1学生,2教职工,3校友)',
    `walk_status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '活动状态(1未开始,2待出发,3进行中,4已放弃,5已下撤,6已违规,7已完成)',
    `created_at`  TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`  TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_people_open_id` (`open_id`),
    UNIQUE KEY `uni_people_identity` (`identity`),
    UNIQUE KEY `uni_people_tel` (`tel`),
    UNIQUE KEY `uni_people_stu_id` (`stu_id`),
    KEY `idx_people_team_id` (`team_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='人员基础信息表';

-- 2. Teams（队伍信息表）
CREATE TABLE `teams` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '队伍ID',
    `name`        VARCHAR(64)     NOT NULL COMMENT '队伍名称',
    `num`         TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '团队人数',
    `password`    VARCHAR(64)     NOT NULL COMMENT '团队加入密码',
    `slogan`      VARCHAR(128)    DEFAULT NULL COMMENT '团队标语',
    `allow_match` TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '是否允许随机匹配',
    `captain`     VARCHAR(64)     NOT NULL COMMENT '队长OpenID',
    `route_id`    BIGINT UNSIGNED DEFAULT 0 COMMENT '团队所属路线ID',
    `point_id`    TINYINT         DEFAULT 0 COMMENT '当前所在点位ID',
    `start_num`   INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '出发时人数',
    `submit`      TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '是否已提交报名',
    `status`      TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '队伍状态(1未出发,2进行中,3已完成,4已下撤)',
    `code`        VARCHAR(128)    DEFAULT NULL COMMENT '签到二维码绑定码',
    `time`        DATETIME(3)     DEFAULT NULL COMMENT '队伍状态更新时间',
    `is_lost`     TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '是否失联',
    `created_at`  TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`  TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_teams_name` (`name`),
    KEY `idx_teams_code` (`code`),
    KEY `idx_teams_route_point` (`route_id`, `point_id`),
    KEY `idx_teams_captain` (`captain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='队伍信息表';

-- 3. Points（点位信息表）
CREATE TABLE `points` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '全局唯一点位ID',
    `cp_id`      BIGINT UNSIGNED NOT NULL COMMENT '校区内点位编号(可跨校区重复)',
    `name`       VARCHAR(64)     DEFAULT NULL COMMENT '点位名称',
    `is_active`  TINYINT(1)      DEFAULT 1 COMMENT '是否启用',
    `created_at` TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_points_cp_id` (`cp_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='点位信息表';

-- 4. Routes（路线配置表）
CREATE TABLE `routes` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '路线ID',
    `code`        VARCHAR(64)     NOT NULL COMMENT '路线代码，如pf-half, pf-full等',
    `name`        VARCHAR(100)    NOT NULL COMMENT '路线名称，如屏峰半程',
    `campus`      TINYINT UNSIGNED NOT NULL COMMENT '所属校区(1朝晖,2屏峰,3莫干山)',
    `is_active`   TINYINT(1)      NOT NULL DEFAULT 1 COMMENT '是否启用',
    `description` VARCHAR(500)    DEFAULT NULL COMMENT '路线描述',
    `created_at`  TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`  TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_routes_code` (`code`),
    KEY `idx_routes_campus` (`campus`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='路线配置表';

-- 5. RouteEdges（路线边关系表）
CREATE TABLE `route_edges` (
    `id`             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '边ID',
    `route_id`       BIGINT UNSIGNED NOT NULL COMMENT '归属路线ID',
    `front_point_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '前一个点位ID',
    `point_id`       BIGINT UNSIGNED DEFAULT NULL COMMENT '当前点位ID',
    `seq`            INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '在路线中的顺序号',
    `created_at`     TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`     TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_route_edges_route_points` (`route_id`, `front_point_id`, `point_id`),
    KEY `idx_route_edges_points` (`front_point_id`, `point_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='路线边关系表';

-- 6. Admins（管理员权限表）
CREATE TABLE `admins` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '管理员ID',
    `open_id`    VARCHAR(128)    DEFAULT NULL COMMENT '微信OpenID',
    `name`       VARCHAR(128)    DEFAULT NULL COMMENT '管理员姓名',
    `account`    VARCHAR(128)    DEFAULT NULL COMMENT '登录账号',
    `password`   VARCHAR(255)    DEFAULT NULL COMMENT '登录密码',
    `permission` VARCHAR(32)     NOT NULL COMMENT '权限级别(super最高权限,manager负责人权限,internal内部权限,external外部权限)',
    `point_name` VARCHAR(128)    DEFAULT NULL COMMENT '负责的点位名称',
    `campus`     VARCHAR(16)     NOT NULL COMMENT '校区(zh朝晖,pf屏峰,mgs莫干山)',
    `created_at` TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_admins_open_id` (`open_id`),
    KEY `idx_admins_point_name` (`point_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='管理员权限表';

-- 7. Checkins（打卡签到记录表）
CREATE TABLE `checkins` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '签到记录ID',
    `admin_id`   BIGINT          DEFAULT NULL COMMENT '签到管理员ID',
    `team_id`    BIGINT          NOT NULL COMMENT '队伍ID',
    `point_id`   TINYINT         DEFAULT NULL COMMENT '签到点位ID',
    `route_id`   BIGINT UNSIGNED NOT NULL COMMENT '路线ID',
    `time`       DATETIME(3)     DEFAULT NULL COMMENT '签到时间',
    `created_at` TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_checkins_team_point` (`team_id`, `point_id`),
    KEY `idx_checkins_route_point` (`route_id`, `point_id`),
    KEY `idx_checkins_time` (`time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='打卡签到记录表';

-- 8. WrongRouteRecords（走错路线记录表）
CREATE TABLE `wrong_route_records` (
    `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `team_id`          BIGINT          NOT NULL COMMENT '队伍ID',
    `origin_route_id`  BIGINT UNSIGNED DEFAULT 0 COMMENT '原正确路线ID',
    `wrong_route_id`   BIGINT UNSIGNED DEFAULT 0 COMMENT '错走的路线ID',
    `admin_id`         BIGINT          DEFAULT NULL COMMENT '记录该情况的管理员ID',
    `remark`           VARCHAR(500)    DEFAULT NULL COMMENT '备注说明',
    `created_time`     DATETIME(3)     NOT NULL COMMENT '记录时间',
    `created_at`       TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`       TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_wrong_route_team` (`team_id`),
    KEY `idx_wrong_route_routes` (`origin_route_id`, `wrong_route_id`),
    KEY `idx_wrong_route_time` (`created_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='走错路线记录表';

-- 9. User（系统用户表 - 遗留/预留表，用于后台登录）
CREATE TABLE `user` (
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `username`   VARCHAR(20)     NOT NULL COMMENT '用户名',
    `password`   VARCHAR(255)    NOT NULL COMMENT '密码',
    `created_at` TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` TIMESTAMP(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='系统用户表';