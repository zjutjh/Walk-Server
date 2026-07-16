CREATE TABLE `messages` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `sender_id` bigint DEFAULT NULL COMMENT '发送者人员ID，为空表示系统消息',
  `receiver_id` bigint NOT NULL COMMENT '接收者人员ID',
  `message` varchar(255) NOT NULL COMMENT '消息内容',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_messages_receiver_id` (`receiver_id`),
  KEY `idx_messages_sender_id` (`sender_id`),
  KEY `idx_messages_created_at` (`created_at`)
);
