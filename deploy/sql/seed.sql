-- 测试种子数据
-- 说明：
-- 1. 这份脚本用固定主键 + ON DUPLICATE KEY UPDATE，重复执行可覆盖同一批测试数据。
-- 2. 管理员明文密码：
--    屏峰线路管理员统一使用：
--      pfadmin123（pf_qd_admin、pf_jls_admin、pf_lmk_admin、pf_cmq_admin、pf_yst_admin、pf_ljs_admin）
--      admin123（pf_pfs_admin、pf_dls_admin、pf_pfsy_admin、pf_zd_admin）
--    其他校区：
--      mgsadmin123（mgs_mid_admin）
--      zhadmin123（zh_mid_admin）

INSERT INTO `routes` (`id`, `name`, `point_name`, `campus`, `is_active`)
VALUES
  (1, 'pf_all',  '屏峰全程',  'pf', 1),
  (2, 'pf_half', '屏峰半程',  'pf', 1),
  (3, 'mgs_all', '莫干山全程', 'mgs', 1),
  (4, 'mgs_half','莫干山半程', 'mgs', 1),
  (5, 'zh',      '朝晖路线',  'zh', 1)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `point_name` = VALUES(`point_name`),
  `campus` = VALUES(`campus`),
  `is_active` = VALUES(`is_active`);

INSERT INTO `points` (`id`, `cp_id`, `name`, `is_active`)
VALUES
  (1,  1, 'qd',      1),
  (2,  2, 'jls',     1),
  (3,  3, 'lmk',     1),
  (4,  4, 'cmq',     1),
  (5,  5, 'yst',     1),
  (6,  6, 'pfs',     1),
  (7,  7, 'dls',     1),
  (8,  8, 'pfsy',    1),
  (9,  9, 'zd',      1),
  (10, 10, 'ljs',    1),
  (11, 101, 'mgs_qd',  1),
  (12, 102, 'mgs_mid', 1),
  (13, 103, 'mgs_zd',  1),
  (14, 201, 'zh_qd',   1),
  (15, 202, 'zh_mid',  1),
  (16, 203, 'zh_zd',   1)
ON DUPLICATE KEY UPDATE
  `cp_id` = VALUES(`cp_id`),
  `name` = VALUES(`name`),
  `is_active` = VALUES(`is_active`);

INSERT INTO `route_edges` (`id`, `prev_point_name`, `point_name`, `route_name`, `seq_order`)
VALUES
  (1,  NULL,  'qd',      'pf_all',  1),
  (2,  'qd',  'jls',     'pf_all',  2),
  (3,  'jls', 'lmk',     'pf_all',  3),
  (4,  'lmk', 'cmq',     'pf_all',  4),
  (5,  'cmq', 'yst',     'pf_all',  5),
  (6,  'yst', 'pfs',     'pf_all',  6),
  (7,  'pfs', 'dls',     'pf_all',  7),
  (8,  'dls', 'pfsy',    'pf_all',  8),
  (9,  'pfsy','zd',      'pf_all',  9),

  (10, NULL,  'qd',      'pf_half', 1),
  (11, 'qd',  'jls',     'pf_half', 2),
  (12, 'jls', 'ljs',     'pf_half', 3),
  (13, 'ljs', 'pfs',     'pf_half', 4),
  (14, 'pfs', 'dls',     'pf_half', 5),
  (15, 'dls', 'pfsy',    'pf_half', 6),
  (16, 'pfsy','zd',      'pf_half', 7),

  (17, NULL,      'mgs_qd',  'mgs_all',  1),
  (18, 'mgs_qd',  'mgs_mid', 'mgs_all',  2),
  (19, 'mgs_mid', 'mgs_zd',  'mgs_all',  3),

  (20, NULL,      'mgs_qd',  'mgs_half', 1),
  (21, 'mgs_qd',  'mgs_zd',  'mgs_half', 2),

  (23, NULL,     'zh_qd',   'zh', 1),
  (24, 'zh_qd',  'zh_mid',  'zh', 2),
  (25, 'zh_mid', 'zh_zd',   'zh', 3)
ON DUPLICATE KEY UPDATE
  `prev_point_name` = VALUES(`prev_point_name`),
  `point_name` = VALUES(`point_name`),
  `route_name` = VALUES(`route_name`),
  `seq_order` = VALUES(`seq_order`);

INSERT INTO `admins` (`id`, `open_id`, `name`, `account`, `password`, `permission`, `point_name`, `campus`)
VALUES
  (1, 'admin_open_pf_qd',   '屏峰起点管理员',   'pf_qd_admin',   '$2a$10$2vPdGh395d6h95FmZGEBdOCfwC2bdTWOJNq/tvM2tsa519ZBxm2gS', 'super', 'qd',      'pf'),
  (2, 'admin_open_pf_jls',  '屏峰金莲寺管理员', 'pf_jls_admin',  '$2a$10$2vPdGh395d6h95FmZGEBdOCfwC2bdTWOJNq/tvM2tsa519ZBxm2gS', 'super', 'jls',     'pf'),
  (3, 'admin_open_pf_lmk',  '屏峰龙门坎管理员', 'pf_lmk_admin',  '$2a$10$2vPdGh395d6h95FmZGEBdOCfwC2bdTWOJNq/tvM2tsa519ZBxm2gS', 'super', 'lmk',     'pf'),
  (4, 'admin_open_pf_cmq',  '屏峰慈母桥管理员', 'pf_cmq_admin',  '$2a$10$2vPdGh395d6h95FmZGEBdOCfwC2bdTWOJNq/tvM2tsa519ZBxm2gS', 'super', 'cmq',     'pf'),
  (5, 'admin_open_pf_yst',  '屏峰元帅亭管理员', 'pf_yst_admin',  '$2a$10$2vPdGh395d6h95FmZGEBdOCfwC2bdTWOJNq/tvM2tsa519ZBxm2gS', 'super', 'yst',     'pf'),
  (6, 'admin_open_pf_ljs',  '屏峰老焦山管理员',   'pf_ljs_admin',  '$2a$10$2vPdGh395d6h95FmZGEBdOCfwC2bdTWOJNq/tvM2tsa519ZBxm2gS', 'super', 'ljs',     'pf'),
  (7, 'admin_open_pf_pfs',  '屏峰山管理员', 'pf_pfs_admin',  '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super',  'pfs',     'pf'),
  (8, 'admin_open_pf_dls',  '屏峰大岭山管理员', 'pf_dls_admin',  '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'dls',     'pf'),
  (9, 'admin_open_pf_pfsy', '屏峰善院管理员',   'pf_pfsy_admin', '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'pfsy',    'pf'),
  (10,'admin_open_pf_zd',   '屏峰终点管理员',   'pf_zd_admin',   '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'zd',      'pf'),
  (11,'admin_open_mgs_mid', '莫干山中段管理员', 'mgs_mid_admin', '$2a$10$Pk6oCYHux6ebsLaW.zLLEO/TwE4xDVZr/iNNoqCiUr6QGoTn4BlWS', 'super', 'mgs_mid', 'mgs'),
  (12,'admin_open_zh_mid',  '朝晖超级管理员',   'zh_mid_admin',  '$2a$10$9g17pt.K/B1WX6gzfXzB7uker9GNHJuqCNYJsiwWSslmEmtN3.DiW', 'super',    'zh_mid',  'zh')
ON DUPLICATE KEY UPDATE
  `open_id` = VALUES(`open_id`),
  `name` = VALUES(`name`),
  `account` = VALUES(`account`),
  `password` = VALUES(`password`),
  `permission` = VALUES(`permission`),
  `point_name` = VALUES(`point_name`),
  `campus` = VALUES(`campus`);

INSERT INTO `teams` (`id`, `name`, `num`, `password`, `slogan`, `allow_match`, `captain`, `submit`, `route_name`, `prev_point_name`, `status`, `is_wrong_route`, `is_reunite`, `code`, `time`, `is_lost`)
VALUES
  (1, '屏峰全程冲线队', 4, 'team123', '全程稳定推进', 1, 'user_open_pf_1', 1, 'pf_all',  'pfs',     'inProgress', 0, 0, 'CODE-PF-001', '2026-05-01 09:30:00.000', 0),
  (2, '屏峰半程完赛队', 4, 'team234', '半程已完赛',   0, 'user_open_pf_5', 1, 'pf_half', 'zd',      'completed',  0, 0, 'CODE-PF-002', '2026-05-01 11:20:00.000', 0),
  (3, '朝晖失联测试队', 3, 'team345', '朝晖联调队',   0, 'user_open_zh_1', 1, 'zh',      'zh_mid',  'inProgress', 0, 0, 'CODE-ZH-001', '2026-05-01 10:10:00.000', 1),
  (4, '屏峰全程未出发队', 4, 'team456', '还没出发',   1, 'user_open_pf_9', 1, 'pf_all',  NULL,      'notStart',   0, 0, NULL,          NULL,                      0),
  (5, '莫干山半程下撤队', 3, 'team567', '安全第一',   0, 'user_open_mgs_1',1, 'mgs_half','mgs_qd',  'withdrawn',  0, 0, 'CODE-MGS-001','2026-05-01 10:50:00.000', 0),
  (6, '莫干山全程走错队', 4, 'team678', '路线有点偏', 0, 'user_open_mgs_4',1, 'mgs_all', 'mgs_mid', 'inProgress', 1, 0, 'CODE-MGS-002','2026-05-01 10:40:00.000', 0)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `num` = VALUES(`num`),
  `password` = VALUES(`password`),
  `slogan` = VALUES(`slogan`),
  `allow_match` = VALUES(`allow_match`),
  `captain` = VALUES(`captain`),
  `submit` = VALUES(`submit`),
  `route_name` = VALUES(`route_name`),
  `prev_point_name` = VALUES(`prev_point_name`),
  `status` = VALUES(`status`),
  `is_wrong_route` = VALUES(`is_wrong_route`),
  `is_reunite` = VALUES(`is_reunite`),
  `code` = VALUES(`code`),
  `time` = VALUES(`time`),
  `is_lost` = VALUES(`is_lost`);

INSERT INTO `peoples` (`id`, `open_id`, `name`, `gender`, `stu_id`, `campus`, `identity`, `role`, `qq`, `wechat`, `college`, `tel`, `created_op`, `join_op`, `team_id`, `type`, `walk_status`)
VALUES
  (1,  'user_open_pf_1',  '张晨', 1, '20260001', 'pf', '330101200001010001', 'captain', '10001', 'wx_pf_1',  '信息学院', '13800000001', 3, 5, 1, 'student', 'inProgress'),
  (2,  'user_open_pf_2',  '李越', 2, '20260002', 'pf', '330101200001010002', 'member',  '10002', 'wx_pf_2',  '信息学院', '13800000002', 3, 5, 1, 'student', 'inProgress'),
  (3,  'user_open_pf_3',  '周航', 1, '20260003', 'pf', '330101200001010003', 'member',  '10003', 'wx_pf_3',  '机械学院', '13800000003', 3, 5, 1, 'student', 'inProgress'),
  (4,  'user_open_pf_4',  '王岚', 2, '20260004', 'pf', '330101200001010004', 'member',  '10004', 'wx_pf_4',  '机械学院', '13800000004', 3, 5, 1, 'student', 'abandoned'),

  (5,  'user_open_pf_5',  '陈宇', 1, '20260005', 'pf', '330101200001010005', 'captain', '10005', 'wx_pf_5',  '电气学院', '13800000005', 3, 5, 2, 'student', 'completed'),
  (6,  'user_open_pf_6',  '何静', 2, '20260006', 'pf', '330101200001010006', 'member',  '10006', 'wx_pf_6',  '电气学院', '13800000006', 3, 5, 2, 'student', 'completed'),
  (7,  'user_open_pf_7',  '孙博', 1, '20260007', 'pf', '330101200001010007', 'member',  '10007', 'wx_pf_7',  '理学院',   '13800000007', 3, 5, 2, 'student', 'completed'),
  (8,  'user_open_pf_8',  '赵敏', 2, '20260008', 'pf', '330101200001010008', 'member',  '10008', 'wx_pf_8',  '理学院',   '13800000008', 3, 5, 2, 'student', 'completed'),

  (9,  'user_open_pf_9',  '徐涛', 1, '20260009', 'pf', '330101200001010009', 'captain', '10009', 'wx_pf_9',  '计算机学院', '13800000009', 3, 5, 4, 'student', 'pending'),
  (10, 'user_open_pf_10', '高宁', 2, '20260010', 'pf', '330101200001010010', 'member',  '10010', 'wx_pf_10', '计算机学院', '13800000010', 3, 5, 4, 'student', 'pending'),
  (11, 'user_open_pf_11', '吴凡', 1, '20260011', 'pf', '330101200001010011', 'member',  '10011', 'wx_pf_11', '材料学院',   '13800000011', 3, 5, 4, 'student', 'pending'),
  (12, 'user_open_pf_12', '冯悦', 2, '20260012', 'pf', '330101200001010012', 'member',  '10012', 'wx_pf_12', '材料学院',   '13800000012', 3, 5, 4, 'student', 'abandoned'),

  (13, 'user_open_zh_1',  '许航', 1, '20260013', 'zh', '330101200001010013', 'captain', '10013', 'wx_zh_1',  '建筑学院', '13800000013', 3, 5, 3, 'teacher', 'inProgress'),
  (14, 'user_open_zh_2',  '彭媛', 2, '20260014', 'zh', '330101200001010014', 'member',  '10014', 'wx_zh_2',  '建筑学院', '13800000014', 3, 5, 3, 'student', 'inProgress'),
  (15, 'user_open_zh_3',  '董杰', 1, '20260015', 'zh', '330101200001010015', 'member',  '10015', 'wx_zh_3',  '法学院',   '13800000015', 3, 5, 3, 'student', 'inProgress'),

  (16, 'user_open_mgs_1', '唐林', 1, '20260016', 'mgs', '330101200001010016', 'captain', '10016', 'wx_mgs_1', '外国语学院', '13800000016', 3, 5, 5, 'teacher', 'withdrawn'),
  (17, 'user_open_mgs_2', '袁雪', 2, '20260017', 'mgs', '330101200001010017', 'member',  '10017', 'wx_mgs_2', '外国语学院', '13800000017', 3, 5, 5, 'student', 'withdrawn'),
  (18, 'user_open_mgs_3', '郭诚', 1, '20260018', 'mgs', '330101200001010018', 'member',  '10018', 'wx_mgs_3', '经管学院',   '13800000018', 3, 5, 5, 'alumnus', 'withdrawn'),

  (19, 'user_open_mgs_4', '谢然', 1, '20260019', 'mgs', '330101200001010019', 'captain', '10019', 'wx_mgs_4', '人文学院', '13800000019', 3, 5, 6, 'student', 'inProgress'),
  (20, 'user_open_mgs_5', '沈瑶', 2, '20260020', 'mgs', '330101200001010020', 'member',  '10020', 'wx_mgs_5', '人文学院', '13800000020', 3, 5, 6, 'student', 'inProgress'),
  (21, 'user_open_mgs_6', '顾晨', 1, '20260021', 'mgs', '330101200001010021', 'member',  '10021', 'wx_mgs_6', '人文学院', '13800000021', 3, 5, 6, 'student', 'violated'),
  (22, 'user_open_mgs_7', '陆芷', 2, '20260022', 'mgs', '330101200001010022', 'member',  '10022', 'wx_mgs_7', '艺术学院', '13800000022', 3, 5, 6, 'student', 'inProgress'),

  (23, 'user_open_free_1', '未组队学生', 1, '20260023', 'pf',  '330101200001010023', 'unbind', '10023', 'wx_free_1', '软件学院', '13800000023', 3, 5, -1, 'student', 'notStart'),
  (24, 'user_open_free_2', '待报名老师', 2, NULL,      'zh',  '330101200001010024', 'unbind', '10024', 'wx_free_2', '教师发展中心', '13800000024', 3, 5, -1, 'teacher', 'notStart')
ON DUPLICATE KEY UPDATE
  `open_id` = VALUES(`open_id`),
  `name` = VALUES(`name`),
  `gender` = VALUES(`gender`),
  `stu_id` = VALUES(`stu_id`),
  `campus` = VALUES(`campus`),
  `identity` = VALUES(`identity`),
  `role` = VALUES(`role`),
  `qq` = VALUES(`qq`),
  `wechat` = VALUES(`wechat`),
  `college` = VALUES(`college`),
  `tel` = VALUES(`tel`),
  `created_op` = VALUES(`created_op`),
  `join_op` = VALUES(`join_op`),
  `team_id` = VALUES(`team_id`),
  `type` = VALUES(`type`),
  `walk_status` = VALUES(`walk_status`);

INSERT INTO `checkins` (`id`, `admin_id`, `team_id`, `point_name`, `route_name`, `time`)
VALUES
  (1, 1, 1, 'qd',      'pf_all',  '2026-05-01 08:10:00.000'),
  (2, 1, 1, 'jls',     'pf_all',  '2026-05-01 08:30:00.000'),
  (3, 3, 1, 'pfs',     'pf_all',  '2026-05-01 09:30:00.000'),
  (4, 1, 2, 'qd',      'pf_half', '2026-05-01 08:00:00.000'),
  (5, 2, 2, 'ljs',     'pf_half', '2026-05-01 09:10:00.000'),
  (6, 4, 2, 'pfsy',    'pf_half', '2026-05-01 10:30:00.000'),
  (7, 6, 3, 'zh_mid',  'zh',      '2026-05-01 10:10:00.000'),
  (8, 11, 5, 'mgs_qd', 'mgs_half','2026-05-01 10:50:00.000'),
  (9, 5, 6, 'mgs_mid', 'mgs_all', '2026-05-01 10:40:00.000')
ON DUPLICATE KEY UPDATE
  `admin_id` = VALUES(`admin_id`),
  `team_id` = VALUES(`team_id`),
  `point_name` = VALUES(`point_name`),
  `route_name` = VALUES(`route_name`),
  `time` = VALUES(`time`);

INSERT INTO `wrong_route_records` (`id`, `team_id`, `route_name`, `wrong_route_name`, `admin_id`)
VALUES
  (1, 6, 'mgs_all', 'mgs_half', 5)
ON DUPLICATE KEY UPDATE
  `team_id` = VALUES(`team_id`),
  `route_name` = VALUES(`route_name`),
  `wrong_route_name` = VALUES(`wrong_route_name`),
  `admin_id` = VALUES(`admin_id`);
