CREATE TABLE iap_transactions (
  id INT AUTO_INCREMENT PRIMARY KEY,
  transaction_id VARCHAR(64) NOT NULL UNIQUE COMMENT '苹果交易号,防重放',
  original_transaction_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  app_name VARCHAR(32) NOT NULL,
  product_id VARCHAR(128) NOT NULL,
  amount DECIMAL(10,2) NOT NULL,
  credits INT NOT NULL,
  environment VARCHAR(16) NOT NULL DEFAULT 'production' COMMENT 'sandbox/production',
  receipt_hash VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'receipt 的 sha256,用于排查',
  pay_time DATETIME NOT NULL,
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user (user_id),
  INDEX idx_original (original_transaction_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Apple IAP 交易记录';
