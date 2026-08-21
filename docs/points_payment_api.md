# 积分支付API文档

## 接口说明

使用积分支付商品，从ProductPrice获取价格，支付完成后在IAPTransaction中记录交易信息。

## API端点

```
POST /v2/points/payment
```

## 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_id | string | 是 | 用户ID |
| app_name | string | 是 | 应用名称 |
| product_name | string | 是 | 商品名称（对应iap_product_prices表的name字段） |

## 请求示例

```json
{
  "user_id": "user123",
  "app_name": "xiaolu",
  "product_name": "credits6"
}
```

## 响应示例

### 成功响应

```json
{
  "code": 200,
  "data": {
    "transaction_id": "POINTS_PAY20250815123456789",
    "product_name": "credits6",
    "credits": 60,
    "balance": 940
  },
  "message": "success"
}
```

### 错误响应

积分不足：
```json
{
  "code": 400,
  "message": "insufficient balance"
}
```

商品不存在：
```json
{
  "code": 404,
  "message": "商品不存在: record not found"
}
```

## 响应字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| transaction_id | string | 交易ID（格式：POINTS_PAY+时间戳） |
| product_name | string | 商品名称 |
| credits | int | 消耗的积分数量 |
| balance | int | 支付后的积分余额 |

## 业务逻辑

1. 根据`product_name`从`iap_product_prices`表查询商品配置
2. 检查用户积分余额是否足够
3. 开启数据库事务：
   - 调用`UserPointsDAO.Decrease()`扣减用户积分
   - 在`point_transactions`表中记录积分变动明细（type=consume）
   - 在`iap_transactions`表中记录交易信息（environment=points）
4. 提交事务并返回结果

## 数据库表

### iap_product_prices
- `name`: 商品内部名称（如credits6）
- `product_id`: App Store商品ID
- `credits`: 需要消耗的积分数量
- `price`: 商品价格（人民币）

### user_points
- `user_id`: 用户ID
- `app_name`: 应用名称
- `balance`: 积分余额

### point_transactions
- `type`: "recharge"(充值) / "consume"(消费) / "refund"(退款)
- `amount`: 积分变动数量（消费为负数）
- `balance_after`: 变动后余额

### iap_transactions
- `transaction_id`: 交易ID
- `environment`: "points"表示积分支付
- `credits`: 消耗的积分数量
