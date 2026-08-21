-- Apple IAP 内购套餐示例数据
-- 上线前请按 App Store Connect 后台实际配置的 product_id 和价格调整
-- product_prices 表的字段说明:
--   name            - 内部命名,只用于业务方识别(如 credits60)
--   price           - 苹果商品的人民币价格(用于落账记录,实际以苹果返回为准)
--   apple_product_id - App Store Connect 里配置的 product_id,必须和苹果后台一致
--   credits         - 该套餐发放的积分数量
--   product_kind    - 商品类型,内购套餐统一为 'credits'

INSERT INTO product_prices (name, price, apple_product_id, credits, product_kind) VALUES
  ('credits6',     6.00,  'cn.xiaolu.credits6',     60,   'credits'),
  ('credits30',   30.00,  'cn.xiaolu.credits30',   300,   'credits'),
  ('credits68',   68.00,  'cn.xiaolu.credits68',   680,   'credits'),
  ('credits128', 128.00,  'cn.xiaolu.credits128', 1280,   'credits'),
  ('credits328', 328.00,  'cn.xiaolu.credits328', 3280,   'credits'),
  ('credits688', 688.00,  'cn.xiaolu.credits688', 6880,   'credits')
ON DUPLICATE KEY UPDATE
  price = VALUES(price),
  apple_product_id = VALUES(apple_product_id),
  credits = VALUES(credits),
  product_kind = VALUES(product_kind);
