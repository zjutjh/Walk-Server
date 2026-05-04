-- 测试种子数据（按《校区 / 路线 / 点位 数据字典》重建）
-- 说明：
-- 1. 这份脚本用固定主键 + ON DUPLICATE KEY UPDATE，重复执行可覆盖同一批测试数据。
-- 2. 当前按数据字典生成：
--    校区：pf、mgs
--    路线：pf-full、pf-half、mgs
--    点位：pfxq、jls、blt、cmq、gzsgy、pfs、pfsy、ljs、mgsxq、zfgy、hbgy、tayg、dtx
-- 3. 管理员明文密码统一为：admin123

INSERT INTO `routes` (`id`, `name`, `point_name`, `campus`, `is_active`)
VALUES
  (1, 'pf-full', '屏峰全程',  'pf',  1),
  (2, 'pf-half', '屏峰半程',  'pf',  1),
  (3, 'mgs',     '莫干山路线', 'mgs', 1)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `point_name` = VALUES(`point_name`),
  `campus` = VALUES(`campus`),
  `is_active` = VALUES(`is_active`);

INSERT INTO `points` (`id`, `cp_id`, `name`, `is_active`)
VALUES
  (1,  1, 'pfxq',  1),
  (2,  2, 'jls',   1),
  (3,  3, 'blt',   1),
  (4,  4, 'cmq',   1),
  (5,  5, 'gzsgy', 1),
  (6,  6, 'pfs',   1),
  (7,  7, 'pfsy',  1),
  (8,  8, 'ljs',   1),
  (9,  1, 'mgsxq', 1),
  (10, 2, 'zfgy',  1),
  (11, 3, 'hbgy',  1),
  (12, 4, 'tayg',  1),
  (13, 5, 'dtx',   1)
ON DUPLICATE KEY UPDATE
  `cp_id` = VALUES(`cp_id`),
  `name` = VALUES(`name`),
  `is_active` = VALUES(`is_active`);

INSERT INTO `route_edges` (`id`, `prev_point_name`, `point_name`, `route_name`, `seq_order`)
VALUES
  (1,  NULL,    'pfxq',  'pf-full', 1),
  (2,  'pfxq',  'jls',   'pf-full', 2),
  (3,  'jls',   'blt',   'pf-full', 3),
  (4,  'blt',   'cmq',   'pf-full', 4),
  (5,  'cmq',   'gzsgy', 'pf-full', 5),
  (6,  'gzsgy', 'pfs',   'pf-full', 6),
  (7,  'pfs',   'pfsy',  'pf-full', 7),
  (8,  'pfsy',  'pfxq',  'pf-full', 8),

  (9,  NULL,    'pfxq',  'pf-half', 1),
  (10, 'pfxq',  'jls',   'pf-half', 2),
  (11, 'jls',   'ljs',   'pf-half', 3),
  (12, 'ljs',   'pfs',   'pf-half', 4),
  (13, 'pfs',   'pfsy',  'pf-half', 5),
  (14, 'pfsy',  'pfxq',  'pf-half', 6),

  (15, NULL,    'mgsxq', 'mgs',     1),
  (16, 'mgsxq', 'zfgy',  'mgs',     2),
  (17, 'zfgy',  'hbgy',  'mgs',     3),
  (18, 'hbgy',  'tayg',  'mgs',     4),
  (19, 'tayg',  'dtx',   'mgs',     5),
  (20, 'dtx',   'mgsxq', 'mgs',     6)
ON DUPLICATE KEY UPDATE
  `prev_point_name` = VALUES(`prev_point_name`),
  `point_name` = VALUES(`point_name`),
  `route_name` = VALUES(`route_name`),
  `seq_order` = VALUES(`seq_order`);

INSERT INTO `admins` (`id`, `open_id`, `name`, `account`, `password`, `permission`, `point_name`, `campus`)
VALUES
  (1,  'admin_open_pf_pfxq',   '屏峰起终点管理员',   'pf_pfxq_admin',   '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'pfxq',  'pf'),
  (2,  'admin_open_pf_jls',    '金莲寺管理员',       'pf_jls_admin',    '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'jls',   'pf'),
  (3,  'admin_open_pf_blt',    '白龙潭管理员',       'pf_blt_admin',    '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'blt',   'pf'),
  (4,  'admin_open_pf_cmq',    '慈母桥管理员',       'pf_cmq_admin',    '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'cmq',   'pf'),
  (5,  'admin_open_pf_gzsgy',  '古樟树公园管理员',   'pf_gzsgy_admin',  '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'gzsgy', 'pf'),
  (6,  'admin_open_pf_pfs',    '屏峰山管理员',       'pf_pfs_admin',    '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'pfs',   'pf'),
  (7,  'admin_open_pf_pfsy',   '屏峰善院管理员',     'pf_pfsy_admin',   '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'pfsy',  'pf'),
  (8,  'admin_open_pf_ljs',    '老焦山管理员',       'pf_ljs_admin',    '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'ljs',   'pf'),
  (9,  'admin_open_mgs_xq',    '莫干山校区起终点管理员', 'mgs_xq_admin',   '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'mgsxq', 'mgs'),
  (10, 'admin_open_mgs_zfgy',  '兆丰公园管理员',         'mgs_zfgy_admin', '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'zfgy',  'mgs'),
  (11, 'admin_open_mgs_hbgy',  '滑板公园管理员',         'mgs_hbgy_admin', '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'hbgy',  'mgs'),
  (12, 'admin_open_mgs_tayg',  '天安云谷管理员',         'mgs_tayg_admin', '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'tayg',  'mgs'),
  (13, 'admin_open_mgs_dtx',   '东苔溪管理员',           'mgs_dtx_admin',  '$2a$10$JPCw.G1REFcorkD50IaoP.o0n8ZWv9vJfRkdVKwG9SVPodTTzVuba', 'super', 'dtx',   'mgs')
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
  (1, '屏峰全程进行中队', 4, 'team123', '全程稳步推进', 1, 'user_open_pf_1',  1, 'pf-full', 'cmq',   'inProgress', 0, 0, 'CODE-PF-001',  '2026-05-01 09:30:00.000', 0),
  (2, '屏峰半程已完赛队', 4, 'team234', '半程顺利收官', 0, 'user_open_pf_5',  1, 'pf-half', 'pfxq',  'completed',  0, 0, 'CODE-PF-002',  '2026-05-01 11:20:00.000', 0),
  (3, '屏峰全程未出发队', 4, 'team345', '等待绑定签到码', 1, 'user_open_pf_9',  1, 'pf-full', NULL,    'notStart',   0, 0, NULL,           NULL,                      0),
  (4, '莫干山失联测试队', 3, 'team456', '莫干山联调',     0, 'user_open_mgs_1', 1, 'mgs',     'hbgy',  'inProgress', 0, 0, 'CODE-MGS-001', '2026-05-01 10:10:00.000', 1),
  (5, '屏峰全程走错队',   4, 'team567', '路线走偏测试',   0, 'user_open_pf_13', 1, 'pf-full', 'jls',   'inProgress', 1, 0, 'CODE-PF-003',  '2026-05-01 08:40:00.000', 0),
  (6, '莫干山下撤测试队', 3, 'team678', '安全第一',       0, 'user_open_mgs_4', 1, 'mgs',     'zfgy',  'withdrawn',  0, 0, 'CODE-MGS-002', '2026-05-01 10:50:00.000', 0),
  (7, '屏峰半程进行中队', 4, 'team789', '半程冲冲冲',     1, 'user_open_pf_17', 1, 'pf-half', 'pfs',   'inProgress', 0, 0, 'CODE-PF-004',  '2026-05-01 10:00:00.000', 0),
  (8, '莫干山完赛测试队', 4, 'team890', '莫干山顺利完赛', 0, 'user_open_mgs_7', 1, 'mgs',     'mgsxq', 'completed',  0, 0, 'CODE-MGS-003', '2026-05-01 12:10:00.000', 0),
  (9, '屏峰全程违规测试队', 4, 'team901', '规则边界测试', 0, 'user_open_pf_21', 1, 'pf-full', 'gzsgy', 'inProgress', 0, 0, 'CODE-PF-005',  '2026-05-01 09:50:00.000', 0),
  (10,'莫干山未出发测试队', 3, 'team012', '等待出发',     1, 'user_open_mgs_11',1, 'mgs',     NULL,    'notStart',   0, 0, NULL,           NULL,                      0)
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
  (1,  'user_open_pf_1',  '张晨', 1, '20260001', 'pf',  '330101200001010001', 'captain', '10001', 'wx_pf_1',  '信息学院',     '13800000001', 3, 5, 1,  'student', 'inProgress'),
  (2,  'user_open_pf_2',  '李越', 2, '20260002', 'pf',  '330101200001010002', 'member',  '10002', 'wx_pf_2',  '信息学院',     '13800000002', 3, 5, 1,  'student', 'inProgress'),
  (3,  'user_open_pf_3',  '周航', 1, '20260003', 'pf',  '330101200001010003', 'member',  '10003', 'wx_pf_3',  '机械学院',     '13800000003', 3, 5, 1,  'student', 'inProgress'),
  (4,  'user_open_pf_4',  '王岚', 2, '20260004', 'pf',  '330101200001010004', 'member',  '10004', 'wx_pf_4',  '机械学院',     '13800000004', 3, 5, 1,  'student', 'abandoned'),

  (5,  'user_open_pf_5',  '陈宇', 1, '20260005', 'pf',  '330101200001010005', 'captain', '10005', 'wx_pf_5',  '电气学院',     '13800000005', 3, 5, 2,  'student', 'completed'),
  (6,  'user_open_pf_6',  '何静', 2, '20260006', 'pf',  '330101200001010006', 'member',  '10006', 'wx_pf_6',  '电气学院',     '13800000006', 3, 5, 2,  'student', 'completed'),
  (7,  'user_open_pf_7',  '孙博', 1, '20260007', 'pf',  '330101200001010007', 'member',  '10007', 'wx_pf_7',  '理学院',       '13800000007', 3, 5, 2,  'student', 'completed'),
  (8,  'user_open_pf_8',  '赵敏', 2, '20260008', 'pf',  '330101200001010008', 'member',  '10008', 'wx_pf_8',  '理学院',       '13800000008', 3, 5, 2,  'student', 'completed'),

  (9,  'user_open_pf_9',  '徐涛', 1, '20260009', 'pf',  '330101200001010009', 'captain', '10009', 'wx_pf_9',  '计算机学院',   '13800000009', 3, 5, 3,  'student', 'pending'),
  (10, 'user_open_pf_10', '高宁', 2, '20260010', 'pf',  '330101200001010010', 'member',  '10010', 'wx_pf_10', '计算机学院',   '13800000010', 3, 5, 3,  'student', 'pending'),
  (11, 'user_open_pf_11', '吴凡', 1, '20260011', 'pf',  '330101200001010011', 'member',  '10011', 'wx_pf_11', '材料学院',     '13800000011', 3, 5, 3,  'student', 'pending'),
  (12, 'user_open_pf_12', '冯悦', 2, '20260012', 'pf',  '330101200001010012', 'member',  '10012', 'wx_pf_12', '材料学院',     '13800000012', 3, 5, 3,  'student', 'abandoned'),

  (13, 'user_open_pf_13', '谢然', 1, '20260013', 'pf',  '330101200001010013', 'captain', '10013', 'wx_pf_13', '人文学院',     '13800000013', 3, 5, 5,  'student', 'inProgress'),
  (14, 'user_open_pf_14', '沈瑶', 2, '20260014', 'pf',  '330101200001010014', 'member',  '10014', 'wx_pf_14', '人文学院',     '13800000014', 3, 5, 5,  'student', 'violated'),
  (15, 'user_open_pf_15', '顾晨', 1, '20260015', 'pf',  '330101200001010015', 'member',  '10015', 'wx_pf_15', '人文学院',     '13800000015', 3, 5, 5,  'student', 'inProgress'),
  (16, 'user_open_pf_16', '陆芷', 2, '20260016', 'pf',  '330101200001010016', 'member',  '10016', 'wx_pf_16', '艺术学院',     '13800000016', 3, 5, 5,  'student', 'inProgress'),

  (17, 'user_open_mgs_1', '许航', 1, '20260017', 'mgs', '330101200001010017', 'captain', '10017', 'wx_mgs_1', '建筑学院',     '13800000017', 3, 5, 4,  'teacher', 'inProgress'),
  (18, 'user_open_mgs_2', '彭媛', 2, '20260018', 'mgs', '330101200001010018', 'member',  '10018', 'wx_mgs_2', '建筑学院',     '13800000018', 3, 5, 4,  'student', 'inProgress'),
  (19, 'user_open_mgs_3', '董杰', 1, '20260019', 'mgs', '330101200001010019', 'member',  '10019', 'wx_mgs_3', '法学院',       '13800000019', 3, 5, 4,  'student', 'inProgress'),

  (20, 'user_open_mgs_4', '唐林', 1, '20260020', 'mgs', '330101200001010020', 'captain', '10020', 'wx_mgs_4', '外国语学院',   '13800000020', 3, 5, 6,  'teacher', 'withdrawn'),
  (21, 'user_open_mgs_5', '袁雪', 2, '20260021', 'mgs', '330101200001010021', 'member',  '10021', 'wx_mgs_5', '外国语学院',   '13800000021', 3, 5, 6,  'student', 'withdrawn'),
  (22, 'user_open_mgs_6', '郭诚', 1, '20260022', 'mgs', '330101200001010022', 'member',  '10022', 'wx_mgs_6', '经管学院',     '13800000022', 3, 5, 6,  'alumnus', 'withdrawn'),

  (23, 'user_open_pf_17', '邵可', 1, '20260023', 'pf',  '330101200001010023', 'captain', '10023', 'wx_pf_17', '软件学院',     '13800000023', 3, 5, 7,  'student', 'inProgress'),
  (24, 'user_open_pf_18', '姜妍', 2, '20260024', 'pf',  '330101200001010024', 'member',  '10024', 'wx_pf_18', '软件学院',     '13800000024', 3, 5, 7,  'student', 'inProgress'),
  (25, 'user_open_pf_19', '罗征', 1, '20260025', 'pf',  '330101200001010025', 'member',  '10025', 'wx_pf_19', '生工学院',     '13800000025', 3, 5, 7,  'student', 'inProgress'),
  (26, 'user_open_pf_20', '韩笑', 2, '20260026', 'pf',  '330101200001010026', 'member',  '10026', 'wx_pf_20', '生工学院',     '13800000026', 3, 5, 7,  'student', 'completed'),

  (27, 'user_open_mgs_7',  '周扬', 1, '20260027', 'mgs', '330101200001010027', 'captain', '10027', 'wx_mgs_7',  '计算机学院',   '13800000027', 3, 5, 8,  'student', 'completed'),
  (28, 'user_open_mgs_8',  '宋依', 2, '20260028', 'mgs', '330101200001010028', 'member',  '10028', 'wx_mgs_8',  '计算机学院',   '13800000028', 3, 5, 8,  'student', 'completed'),
  (29, 'user_open_mgs_9',  '曾轩', 1, '20260029', 'mgs', '330101200001010029', 'member',  '10029', 'wx_mgs_9',  '设计学院',     '13800000029', 3, 5, 8,  'student', 'completed'),
  (30, 'user_open_mgs_10', '钱宁', 2, '20260030', 'mgs', '330101200001010030', 'member',  '10030', 'wx_mgs_10', '设计学院',     '13800000030', 3, 5, 8,  'student', 'completed'),

  (31, 'user_open_pf_21', '严哲', 1, '20260031', 'pf',  '330101200001010031', 'captain', '10031', 'wx_pf_21', '自动化学院',   '13800000031', 3, 5, 9,  'student', 'inProgress'),
  (32, 'user_open_pf_22', '孔岚', 2, '20260032', 'pf',  '330101200001010032', 'member',  '10032', 'wx_pf_22', '自动化学院',   '13800000032', 3, 5, 9,  'student', 'violated'),
  (33, 'user_open_pf_23', '蒋程', 1, '20260033', 'pf',  '330101200001010033', 'member',  '10033', 'wx_pf_23', '土木学院',     '13800000033', 3, 5, 9,  'student', 'inProgress'),
  (34, 'user_open_pf_24', '贺青', 2, '20260034', 'pf',  '330101200001010034', 'member',  '10034', 'wx_pf_24', '土木学院',     '13800000034', 3, 5, 9,  'student', 'abandoned'),

  (35, 'user_open_mgs_11', '苏禾', 1, '20260035', 'mgs', '330101200001010035', 'captain', '10035', 'wx_mgs_11', '数学学院',     '13800000035', 3, 5, 10, 'student', 'pending'),
  (36, 'user_open_mgs_12', '陶然', 2, '20260036', 'mgs', '330101200001010036', 'member',  '10036', 'wx_mgs_12', '数学学院',     '13800000036', 3, 5, 10, 'student', 'pending'),
  (37, 'user_open_mgs_13', '叶菲', 2, '20260037', 'mgs', '330101200001010037', 'member',  '10037', 'wx_mgs_13', '金融学院',     '13800000037', 3, 5, 10, 'student', 'abandoned'),

  (38, 'user_open_free_1', '未组队学生', 1, '20260038', 'pf',  '330101200001010038', 'unbind', '10038', 'wx_free_1', '软件学院',     '13800000038', 3, 5, -1, 'student', 'notStart'),
  (39, 'user_open_free_2', '待报名老师', 2, NULL,       'mgs', '330101200001010039', 'unbind', '10039', 'wx_free_2', '教师发展中心', '13800000039', 3, 5, -1, 'teacher', 'notStart')
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
  (1, 1, 1, 'pfxq',  'pf-full', '2026-05-01 08:10:00.000'),
  (2, 2, 1, 'jls',   'pf-full', '2026-05-01 08:30:00.000'),
  (3, 3, 1, 'blt',   'pf-full', '2026-05-01 09:00:00.000'),
  (4, 4, 1, 'cmq',   'pf-full', '2026-05-01 09:30:00.000'),
  (5, 1, 2, 'pfxq',  'pf-half', '2026-05-01 08:00:00.000'),
  (6, 2, 2, 'jls',   'pf-half', '2026-05-01 08:40:00.000'),
  (7, 8, 2, 'ljs',   'pf-half', '2026-05-01 09:10:00.000'),
  (8, 7, 2, 'pfsy',  'pf-half', '2026-05-01 10:30:00.000'),
  (9, 1, 2, 'pfxq',  'pf-half', '2026-05-01 11:20:00.000'),
  (10, 9,  4, 'mgsxq', 'mgs',     '2026-05-01 09:20:00.000'),
  (11, 10, 4, 'zfgy',  'mgs',     '2026-05-01 09:45:00.000'),
  (12, 11, 4, 'hbgy',  'mgs',     '2026-05-01 10:10:00.000'),
  (13, 2,  5, 'jls',   'pf-full', '2026-05-01 08:40:00.000'),
  (14, 9,  6, 'mgsxq', 'mgs',     '2026-05-01 10:20:00.000'),
  (15, 10, 6, 'zfgy',  'mgs',     '2026-05-01 10:50:00.000'),
  (16, 1,  7, 'pfxq',  'pf-half', '2026-05-01 08:30:00.000'),
  (17, 2,  7, 'jls',   'pf-half', '2026-05-01 09:00:00.000'),
  (18, 6,  7, 'pfs',   'pf-half', '2026-05-01 10:00:00.000'),
  (19, 9,  8, 'mgsxq', 'mgs',     '2026-05-01 08:50:00.000'),
  (20, 10, 8, 'zfgy',  'mgs',     '2026-05-01 09:20:00.000'),
  (21, 11, 8, 'hbgy',  'mgs',     '2026-05-01 10:00:00.000'),
  (22, 12, 8, 'tayg',  'mgs',     '2026-05-01 10:40:00.000'),
  (23, 13, 8, 'dtx',   'mgs',     '2026-05-01 11:20:00.000'),
  (24, 9,  8, 'mgsxq', 'mgs',     '2026-05-01 12:10:00.000'),
  (25, 1,  9, 'pfxq',  'pf-full', '2026-05-01 08:20:00.000'),
  (26, 2,  9, 'jls',   'pf-full', '2026-05-01 08:50:00.000'),
  (27, 3,  9, 'blt',   'pf-full', '2026-05-01 09:10:00.000'),
  (28, 4,  9, 'cmq',   'pf-full', '2026-05-01 09:30:00.000'),
  (29, 5,  9, 'gzsgy', 'pf-full', '2026-05-01 09:50:00.000')
ON DUPLICATE KEY UPDATE
  `admin_id` = VALUES(`admin_id`),
  `team_id` = VALUES(`team_id`),
  `point_name` = VALUES(`point_name`),
  `route_name` = VALUES(`route_name`),
  `time` = VALUES(`time`);

INSERT INTO `wrong_route_records` (`id`, `team_id`, `route_name`, `wrong_route_name`, `admin_id`)
VALUES
  (1, 5, 'pf-full', 'pf-half', 8)
ON DUPLICATE KEY UPDATE
  `team_id` = VALUES(`team_id`),
  `route_name` = VALUES(`route_name`),
  `wrong_route_name` = VALUES(`wrong_route_name`),
  `admin_id` = VALUES(`admin_id`);
