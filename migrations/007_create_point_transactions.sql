CREATE TABLE point_transactions (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  app_name VARCHAR(32) NOT NULL DEFAULT '',
  type VARCHAR(16) NOT NULL COMMENT 'recharge/consume/refund',
  amount INT NOT NULL COMMENT '正数为增加,负数为扣减',
  balance_after INT NOT NULL,
  source VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'iap/alipay/wechat/manual/admin',
  ref_order_no VARCHAR(64) NOT NULL DEFAULT '' COMMENT '关联订单号(IAP 用 original_transaction_id)',
  product_id VARCHAR(128) NOT NULL DEFAULT '',
  remark VARCHAR(255) NOT NULL DEFAULT '',
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user (user_id),
  INDEX idx_ref (ref_order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='积分变动明细';
