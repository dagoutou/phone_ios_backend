# 短信支付 API 使用示例

## 前提条件

1. 用户已注册并有足够余额
2. 已配置短信 API 参数
3. 已创建数据库表

## API 接口详情

### 1. 发送短信（直接支付）

**接口地址：** `POST /v2/sms/send`

**请求参数：**

| 参数名 | 类型 | 必选 | 说明 |
|--------|------|------|------|
| user_id | string | 是 | 用户ID |
| amount | string | 是 | 支付金额 |
| content | string | 是 | 短信内容 |
| mobile | string | 是 | 收件人号码 |
| account | string | 否 | 发件人账号（必须传验证过的真实号码） |
| sms_autograph | string | 是 | 短信签名ID |
| houzhui | string | 否 | 短信后缀 |
| plan_time | int64 | 否 | 定时发送时间（秒级时间戳） |
| app_name | string | 是 | 应用名称（如：xiaolu） |

**响应示例：**

```json
{
  "code": 200,
  "data": {
    "status": "success",
    "order_no": "Balance20240101120000123456"
  },
  "message": "success"
}
```

**curl 示例：**

```bash
curl -X POST http://localhost:8088/v2/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test_user_001",
    "amount": "0.1",
    "content": "您的验证码是123456",
    "mobile": "13800138000",
    "account": "13900139000",
    "sms_autograph": "102960",
    "app_name": "xiaolu"
  }'
```

### 2. 通用支付接口（支持多种商品类型）

**接口地址：** `POST /v2/payments`

**请求参数：**

| 参数名 | 类型 | 必选 | 说明 |
|--------|------|------|------|
| user_id | string | 是 | 用户ID |
| amount | string | 是 | 支付金额 |
| payment_type | string | 是 | 支付类型（wechat/alipay/balance） |
| product_type | string | 是 | 商品类型（phone/recharge/text_message） |
| app_name | string | 是 | 应用名称 |
| metadata | string | 是 | 商品参数JSON字符串 |
| description | string | 否 | 描述 |

**短信商品 metadata 参数：**

```json
{
  "content": "短信内容",
  "mobile": "收件人号码",
  "account": "发件人号码",
  "sms_autograph": "短信签名ID",
  "houzhui": "短信后缀",
  "plan_time": 1659664607
}
```

**完整请求示例：**

```bash
curl -X POST http://localhost:8088/v2/payments \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test_user_001",
    "amount": "0.1",
    "payment_type": "balance",
    "product_type": "text_message",
    "app_name": "xiaolu",
    "description": "发送短信",
    "metadata": "{\"content\":\"您的验证码是123456\",\"mobile\":\"13800138000\",\"account\":\"13900139000\",\"sms_autograph\":\"102960\"}"
  }'
```

### 3. 查询短信订单列表

**接口地址：** `GET /v2/sms/orders`

**请求参数：**

| 参数名 | 类型 | 必选 | 说明 |
|--------|------|------|------|
| user_id | string | 是 | 用户ID |

**请求示例：**

```bash
curl -X GET "http://localhost:8088/v2/sms/orders?user_id=test_user_001"
```

**响应示例：**

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "order_no": "Balance20240101120000123456",
      "user_id": "test_user_001",
      "order_type": "text_message",
      "status": "paid",
      "plan_time": null,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "message": "success"
}
```

### 4. 查询短信详情列表

**接口地址：** `GET /v2/sms/details`

**请求参数：**

| 参数名 | 类型 | 必选 | 说明 |
|--------|------|------|------|
| user_id | string | 是 | 用户ID |

**请求示例：**

```bash
curl -X GET "http://localhost:8088/v2/sms/details?user_id=test_user_001"
```

**响应示例：**

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "api_send_id": "2208050956474807877",
      "user_id": "test_user_001",
      "order_no": "Balance20240101120000123456",
      "mobile": "13800138000",
      "account": "13900139000",
      "content": "您的验证码是123456",
      "plan_time": null,
      "create_time": "2024-01-01T12:00:00Z",
      "strip": 1,
      "remarks": "已提交",
      "domain": "yuming.top",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "message": "success"
}
```

### 5. 查询单个短信订单详情

**接口地址：** `GET /v2/sms/order/detail`

**请求参数：**

| 参数名 | 类型 | 必选 | 说明 |
|--------|------|------|------|
| order_no | string | 是 | 订单号 |

**请求示例：**

```bash
curl -X GET "http://localhost:8088/v2/sms/order/detail?order_no=Balance20240101120000123456"
```

### 6. 接收短信上行回复（短信系统回调）

**接口地址：** `POST /v2/pull_sms_upstream`

此接口由短信系统调用，不需要用户认证。只支持 HTTP 协议。

**请求参数：**

| 参数名 | 类型 | 必选 | 说明 |
|--------|------|------|------|
| domain | string | 否 | 授权域名 |
| content | string | 是 | 上行回复内容 |
| mobile | string | 是 | 回复的手机号码 |
| api_send_id | string | 是 | 发送的唯一ID，与发短信时返回的对应 |
| create_time | string | 是 | 上行回复时间，格式：YYYY-MM-DD HH:MM:SS |
| lang | string | 否 | 系统语言 |

**短信系统回调示例：**

```bash
curl -X POST http://localhost:8088/v2/pull_sms_upstream \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "yuming.top",
    "content": "测试回复",
    "mobile": "18888888888",
    "api_send_id": "2208050956474807877",
    "create_time": "2022-08-04 17:34:51",
    "lang": "cn"
  }'
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success"
}
```

**处理流程：**
1. 接收短信系统的回调请求
2. 根据 api_send_id 查找原短信记录
3. 在 text_message_reply 表中创建回复记录
4. 将 text_message_details 表中对应记录的 remarks 更新为 "success"

### 7. 接收短信状态推送（短信系统回调）

**接口地址：** `POST /v2/pull_sms_status`

此接口由短信系统调用，用于接收短信发送失败的状态推送。不需要用户认证。只支持 HTTP 协议。

**请求参数：**

| 参数名 | 类型 | 必选 | 说明 |
|--------|------|------|------|
| domain | string | 否 | 授权域名 |
| send_id | string | 否 | 发送编号（暂时没用，可忽略） |
| remarks | string | 是 | 发送失败的提示信息 |
| account | string | 否 | 发送人账号 |
| api_send_id | string | 是 | 发送的唯一ID，与发短信时返回的对应 |
| lang | string | 否 | 系统语言 |

**短信系统回调示例：**

```bash
curl -X POST http://localhost:8088/v2/pull_sms_status \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "yuming.top",
    "send_id": "36409095",
    "remarks": "驳回",
    "account": "oxmYc5jaVCj4Vmc9p1AHG-NmUCQQ",
    "api_send_id": "2208041724054989486",
    "lang": "cn"
  }'
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success"
}
```

**处理流程：**
1. 接收短信系统的状态推送请求
2. 根据 api_send_id 查找原短信记录
3. 将 text_message_details 表中对应记录的 remarks 更新为 "failed"

### 8. 查询短信回复列表

**接口地址：** `GET /v2/sms/replies`

**请求参数：**

| 参数名 | 类型 | 必选 | 说明 |
|--------|------|------|------|
| user_id | string | 是 | 用户ID |

**请求示例：**

```bash
curl -X GET "http://localhost:8088/v2/sms/replies?user_id=test_user_001"
```

**响应示例：**

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "api_send_id": "2208050956474807877",
      "user_id": "test_user_001",
      "order_no": "Balance20240101120000123456",
      "domain": "yuming.top",
      "content": "测试回复",
      "mobile": "18888888888",
      "create_time": "2022-08-04T17:34:51Z",
      "lang": "cn",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ],
  "message": "success"
}
```

### 9. 查询订单的短信回复

**接口地址：** `GET /v2/sms/replies/by_order`

**请求参数：**

| 参数名 | 类型 | 必选 | 说明 |
|--------|------|------|------|
| order_no | string | 是 | 订单号 |

**请求示例：**

```bash
curl -X GET "http://localhost:8088/v2/sms/replies/by_order?order_no=Balance20240101120000123456"
```

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 400 | 余额不足 |
| 500 | 短信发送失败 |
| 500 | 服务器内部错误 |

## 定时发送示例

```bash
# 计划在 2024-01-01 13:00:00 发送
# 时间戳计算：date -d "2024-01-01 13:00:00" +%s
timestamp=1704118800

curl -X POST http://localhost:8088/v2/sms/send \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"test_user_001\",
    \"amount\": \"0.1\",
    \"content\": \"定时短信测试\",
    \"mobile\": \"13800138000\",
    \"sms_autograph\": \"102960\",
    \"plan_time\": $timestamp,
    \"app_name\": \"xiaolu\"
  }"
```

## 批量发送示例

```bash
# 批量发送多条短信
for phone in "13800138000" "13900139000" "13700137000"; do
  curl -X POST http://localhost:8088/v2/sms/send \
    -H "Content-Type: application/json" \
    -d "{
      \"user_id\": \"test_user_001\",
      \"amount\": \"0.1\",
      \"content\": \"批量发送测试\",
      \"mobile\": \"$phone\",
      \"sms_autograph\": \"102960\",
      \"app_name\": \"xiaolu\"
    }"
  echo "Sent to $phone"
done
```

## 注意事项

1. **余额检查**：发送前请确保用户有足够余额
2. **真实号码**：发件人账号必须使用已验证的真实号码
3. **签名配置**：短信签名ID需要提前配置
4. **内容审核**：短信内容需符合相关规定
5. **定时发送**：定时时间使用秒级时间戳
6. **费率说明**：每条短信按实际条数计费（strip 字段）
