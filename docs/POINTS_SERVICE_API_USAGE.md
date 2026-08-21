# 积分支付业务接口对接文档

本文档用于前端对接以下两个使用积分支付的业务接口：

| 接口 | 用途 |
|---|---|
| `POST /v2/phone_check` | 手机号在网状态查询 |
| `POST /v2/sms/send` | 发送短信或提交人工短信审核 |

## 通用说明

- 请求方式：`POST`
- 请求头：`Content-Type: application/json`
- `user_id`、`app_name` 用于确定扣减哪个用户、哪个应用的积分账户。
- 接口会先扣减积分并记录交易，再执行具体业务。
- `credits` 表示本次实际扣减的积分数量，`balance` 表示扣减后的积分余额。
- 商品价格由服务端配置，前端不要自行传金额，也不要根据本地金额计算积分。
- 当前 `/v2` 路由不强制要求 JWT Token；如服务端后续增加鉴权，前端统一补充 `Authorization: Bearer <token>`。

通用成功响应格式：

```json
{
  "code": 200,
  "data": {},
  "message": "success"
}
```

常见错误响应：

```json
{
  "code": 400,
  "message": "insufficient balance"
}
```

前端应以响应体中的 `code` 判断业务是否成功，不要只判断 HTTP 状态码。

## 1. 手机号在网状态查询

### 请求

接口：`POST /v2/phone_check`

| 参数 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| `user_id` | string | 是 | 用户 ID |
| `app_name` | string | 是 | 应用名称，例如 `xiaolu` |
| `phone` | string | 是 | 待查询的手机号 |

请求示例：

```bash
curl -X POST "https://your-domain.com/v2/phone_check" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_001",
    "app_name": "xiaolu",
    "phone": "13800138000"
  }'
```

### 成功响应

```json
{
  "code": 200,
  "data": {
    "status": "success",
    "transaction_id": "POINTS_PAY20260821123456789",
    "credits": 3,
    "balance": 97,
    "phone_check": {
      "phone": "13800138000",
      "status": "正常",
      "isp": "中国移动",
      "is_xhzw": "否",
      "isp_real": "中国移动"
    }
  },
  "message": "success"
}
```

`phone_check` 字段说明：

| 字段 | 类型 | 说明 |
|---|---|---|
| `phone` | string | 查询手机号 |
| `status` | string | 手机号在网状态 |
| `isp` | string | 运营商信息 |
| `is_xhzw` | string | 是否虚拟运营商，具体值以服务端返回为准 |
| `isp_real` | string | 实际运营商 |

### 前端处理建议

1. 调用接口前可先展示当前积分余额。
2. 返回 `code: 200` 后展示 `phone_check` 查询结果，并刷新积分余额为 `data.balance`。
3. 返回 `code: 400` 且提示积分不足时，引导用户充值。
4. `transaction_id` 用于客服或问题排查，前端无需再次提交。

## 2. 短信发送/人工短信审核

### 请求

接口：`POST /v2/sms/send`

| 参数 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| `user_id` | string | 是 | 用户 ID |
| `app_name` | string | 是 | 应用名称，例如 `xiaolu` |
| `content` | string | 是 | 短信内容 |
| `mobile` | string | 是 | 接收手机号 |
| `message_type` | string | 否 | `text_message` 或 `manual_message`，默认 `text_message` |
| `account` | string | 否 | 发件账号/手机号 |
| `houzhui` | string | 否 | 短信后缀 |
| `plan_time` | int64 | 否 | 定时发送时间，Unix 秒级时间戳 |
| `platform` | string | 条件必填 | `manual_message` 时必填，传话平台 |
| `platform_account` | string | 条件必填 | `manual_message` 时必填，传话账号 |

注意：请求中不需要传 `amount`、`payment_type` 或支付订单号，服务端会根据 `message_type` 查询商品价格并扣减积分。

### 2.1 直接发送短信

请求示例：

```bash
curl -X POST "https://your-domain.com/v2/sms/send" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_001",
    "app_name": "xiaolu",
    "content": "您的验证码是 123456",
    "mobile": "13800138000",
    "account": "13900139000",
    "message_type": "text_message"
  }'
```

成功后服务端会扣减 `text_message` 商品对应积分，并调用短信发送服务。

### 2.2 提交人工短信审核

请求示例：

```bash
curl -X POST "https://your-domain.com/v2/sms/send" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_001",
    "app_name": "xiaolu",
    "content": "请协助联系客户",
    "mobile": "13800138000",
    "message_type": "manual_message",
    "platform": "微信",
    "platform_account": "customer_service_001"
  }'
```

`manual_message` 缺少 `platform` 或 `platform_account` 时，请求会失败且不会扣积分。

### 成功响应

```json
{
  "code": 200,
  "data": {
    "status": "success",
    "order_no": "POINTS_PAY20260821123456789",
    "transaction_id": "POINTS_PAY20260821123456789",
    "credits": 3,
    "balance": 94
  },
  "message": "success"
}
```

字段说明：

| 字段 | 类型 | 说明 |
|---|---|---|
| `status` | string | 成功时为 `success` |
| `order_no` | string | 短信订单号，同时也是本次积分交易号 |
| `transaction_id` | string | 积分交易号，与 `order_no` 相同 |
| `credits` | int | 本次扣减积分 |
| `balance` | int | 扣减后的积分余额 |

### 订单查询

短信支付成功后，可使用返回的 `order_no` 查询订单：

```http
GET /v2/sms/order/detail?order_no=POINTS_PAY20260821123456789
```

也可以查询用户的订单列表和短信详情：

```http
GET /v2/sms/orders?user_id=user_001
GET /v2/sms/details?user_id=user_001&message_type=text_message
```

## 常见错误

| `code` | 含义 | 前端处理建议 |
|---:|---|---|
| `400` | 参数缺失、积分不足、商品配置无效 | 提示用户修正参数或充值 |
| `404` | 商品不存在 | 联系服务端检查商品配置 |
| `500` / `5001` | 服务端或第三方服务异常 | 提示稍后重试，并保留 `transaction_id` 便于排查 |

手机号查询或短信发送属于第三方业务调用；如果积分已经扣除但后续业务调用失败，请将返回信息和交易号提供给服务端处理。
