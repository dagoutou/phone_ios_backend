ALTER TABLE product_prices
  ADD COLUMN apple_product_id VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'App Store 内购商品 ID',
  ADD COLUMN credits INT NOT NULL DEFAULT 0 COMMENT '充值赠送的积分数',
  ADD COLUMN product_kind VARCHAR(32) NOT NULL DEFAULT 'phone_check' COMMENT 'phone_check/credits',
  ADD INDEX idx_apple_pid (apple_product_id);
