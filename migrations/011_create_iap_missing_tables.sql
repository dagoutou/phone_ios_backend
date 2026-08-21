-- 手动创建 IAP 相关表（如果 AutoMigrate 失败可使用此脚本）
-- 执行方式: mysql -u root -p app_phone_andro < migrations/011_create_iap_missing_tables.sql

-- 1. 创建 iap_transactions 表
CREATE TABLE IF NOT EXISTS `iap_transactions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `transaction_id` varchar(255) NOT NULL COMMENT '交易ID',
  `original_transaction_id` varchar(255) DEFAULT NULL COMMENT '原始交易ID',
  `user_id` varchar(255) NOT NULL COMMENT '用户ID',
  `app_name` varchar(255) NOT NULL COMMENT '应用名称',
  `product_id` varchar(255) NOT NULL COMMENT '产品ID',
  `amount` double DEFAULT NULL COMMENT '金额',
  `credits` int NOT NULL COMMENT '积分数量',
  `environment` varchar(50) DEFAULT NULL COMMENT '环境(sandbox/production)',
  `receipt_hash` varchar(255) DEFAULT NULL COMMENT '收据哈希',
  `pay_time` datetime DEFAULT NULL COMMENT '支付时间',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_transaction_id` (`transaction_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_app_name` (`app_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Apple IAP 交易记录';

-- 2. 创建 user_points 表
CREATE TABLE IF NOT EXISTS `user_points` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(255) NOT NULL COMMENT '用户ID',
  `app_name` varchar(255) NOT NULL COMMENT '应用名称',
  `balance` int NOT NULL DEFAULT '0' COMMENT '积分余额',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_app` (`user_id`, `app_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户积分余额';

-- 3. 创建 point_transactions 表
CREATE TABLE IF NOT EXISTS `point_transactions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` varchar(255) NOT NULL COMMENT '用户ID',
  `app_name` varchar(255) NOT NULL COMMENT '应用名称',
  `type` varchar(50) NOT NULL COMMENT '类型(recharge/consume/refund)',
  `amount` int NOT NULL COMMENT '积分变动数量',
  `balance_after` int NOT NULL COMMENT '变动后余额',
  `source` varchar(50) DEFAULT NULL COMMENT '来源(iap/alipay/wechat/manual/admin)',
  `ref_order_no` varchar(255) DEFAULT NULL COMMENT '关联订单号',
  `product_id` varchar(255) DEFAULT NULL COMMENT '产品ID',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_app_name` (`app_name`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分变动明细';
