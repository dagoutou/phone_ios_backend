# 短信类型商品支付流程实现文档

## 概述

本文档描述了短信类型商品支付流程的完整实现，包括数据库设计、模型层、服务层、API 层和路由配置。

## 架构设计

### 1. 数据库表设计

#### orders 表
存储短信订单信息，包含订单号、用户ID、订单类型、状态等基本信息。

```sql
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_no VARCHAR(64) UNIQUE NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    order_type VARCHAR(20) NOT NULL COMMENT 'text_message for SMS orders',
    status VARCHAR(20) NOT NULL DEFAULT 'paid',
    plan_time BIGINT COMMENT 'Scheduled send time (timestamp)',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
```

#### text_message_details 表
存储短信发送详情，包含API返回的完整信息。

```sql
CREATE TABLE IF NOT EXISTS text_message_details (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    api_send_id VARCHAR(50),
    user_id VARCHAR(64) NOT NULL,
    order_no VARCHAR(64) UNIQUE NOT NULL,
    mobile VARCHAR(20) NOT NULL,
    account VARCHAR(20),
    content TEXT,
    plan_time DATETIME,
    create_time DATETIME,
    strip INT,
    remarks VARCHAR(50),
    domain VARCHAR(100),
    created_at DATETIME NOT NULL
);
```

#### text_message_reply 表
存储短信上行回复信息。

```sql
CREATE TABLE IF NOT EXISTS text_message_reply (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    api_send_id VARCHAR(50) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    order_no VARCHAR(64) NOT NULL,
    domain VARCHAR(100),
    content TEXT NOT NULL,
    mobile VARCHAR(20) NOT NULL,
    create_time DATETIME,
    lang VARCHAR(10),
    created_at DATETIME NOT NULL
);
```

### 2. 支付流程

#### 2.1 调用支付接口
```
POST /v2/payments
```

请求参数：
```json
{
  "user_id": "user123",
  "amount": "0.1",
  "payment_type": "balance",
  "product_type": "text_message",
  "app_name": "xiaolu",
  "metadata": "{\"content\":\"短信内容\",\"mobile\":\"13800138000\",\"sms_autograph\":\"102960\",\"account\":\"发件人号码\"}"
}
```

#### 2.2 支付处理流程
1. 验证用户余额是否足够
2. 扣减用户余额
3. 创建支付订单（status = paid）
4. 根据商品类型处理后续业务
   - text_message: 调用短信发送服务
   - phone: 调用电话订单服务
   - recharge: 调用充值服务

#### 2.3 短信发送流程
1. 创建短信订单（orders 表）
2. 调用第三方短信 API
3. 存储短信发送详情（text_message_details 表）
4. 返回支付结果

### 3. 配置文件

在 `config/config.yaml` 中添加短信 API 配置：

```yaml
SMS:
  SignName: "深圳市逐鹿网络"
  TemplateCode: "SMS_293635313"
  Count: 5
  Interval: 5
  # Anonymous SMS API config
  APIDomain: "your-domain.com"
  APIAccount: "your_account"
  APIPassword: "your_password"
  APIProduct: "6"
  SendURL: "http://接口域名/send_sms"
```

### 4. API 接口

#### 4.1 发送短信
```
POST /v2/sms/send
```

请求参数：
```json
{
  "user_id": "user123",
  "amount": "0.1",
  "content": "短信内容",
  "mobile": "13800138000",
  "account": "发件人号码",
  "sms_autograph": "102960",
  "app_name": "xiaolu",
  "plan_time": 1659664607
}
```

#### 4.2 获取短信订单列表
```
GET /v2/sms/orders?user_id=user123
```

#### 4.3 获取短信详情列表
```
GET /v2/sms/details?user_id=user123
```

#### 4.4 获取短信订单详情
```
GET /v2/sms/order/detail?order_no=Balance20240101120000123456
```

#### 4.5 接收短信上行回复
```
POST /v2/pull_sms_upstream
```

此接口由短信系统调用，用于接收用户的上行回复。

**请求参数：**
```json
{
  "domain": "yuming.top",
  "content": "测试",
  "mobile": "18888888888",
  "api_send_id": "2208050956474807877",
  "create_time": "2022-08-04 17:34:51",
  "lang": "cn"
}
```

**处理逻辑：**
1. 根据 api_send_id 查找原短信详情
2. 创建 text_message_reply 记录
3. 更新 text_message_details 的 remarks 为 "success"

#### 4.6 接收短信状态推送
```
POST /v2/pull_sms_status
```

此接口由短信系统调用，用于接收短信发送失败的状态推送。

**请求参数：**
```json
{
  "domain": "yuming.top",
  "send_id": "36409095",
  "remarks": "驳回",
  "account": "oxmYc5jaVCj4Vmc9p1AHG-NmUCQQ",
  "api_send_id": "2208041724054989486",
  "lang": "cn"
}
```

**处理逻辑：**
1. 根据 api_send_id 查找原短信详情
2. 更新 text_message_details 的 remarks 为 "failed"

#### 4.7 获取短信回复列表
```
GET /v2/sms/replies?user_id=user123
```

#### 4.8 获取订单的短信回复
```
GET /v2/sms/replies/by_order?order_no=Balance20240101120000123456
```

### 5. 文件结构

```
andro/
├── migrations/
│   ├── 001_create_orders_table.sql
│   ├── 002_create_text_message_details_table.sql
│   └── 003_create_text_message_reply_table.sql
├── model/
│   └── text_message_order.go
├── dao/
│   └── text_message_order.go
├── services/
│   └── sms/
│       ├── text_message_service.go
│       └── text_message_pay_service.go
├── api/
│   ├── text_message.go
│   └── sms_callback.go
├── config/
│   ├── config.go
│   └── config.yaml
└── services/pay/
    └── balance_pay.go
```

### 6. 数据流图

```
用户请求
  ↓
/v2/payments (支付接口)
  ↓
BalancePayService.CreatePayment()
  ↓
扣减余额
  ↓
创建支付订单
  ↓
检查商品类型 (product_type)
  ↓
text_message → ProcessTextMessagePayment()
  ↓
创建短信订单
  ↓
调用第三方短信 API
  ↓
存储短信详情
  ↓
返回支付结果
```

### 6.1 短信上行回复流程

```
用户回复短信
  ↓
短信系统处理
  ↓
POST /v2/pull_sms_upstream (回调接口)
  ↓
PullSMSUpstream()
  ↓
根据 api_send_id 查询原短信记录
  ↓
创建 text_message_reply 记录
  ↓
更新 text_message_details.remarks = "success"
  ↓
返回成功响应
```

### 6.2 短信状态推送流程

```
短信发送失败
  ↓
短信系统处理
  ↓
POST /v2/pull_sms_status (回调接口)
  ↓
PullSMSStatus()
  ↓
根据 api_send_id 查询原短信记录
  ↓
更新 text_message_details.remarks = "failed"
  ↓
返回成功响应
```

### 7. 错误处理

- 余额不足：返回 400 错误
- 短信发送失败：更新订单状态为 failed
- 数据库操作失败：回滚事务
- API 调用失败：记录日志并返回错误信息

### 8. 测试用例

#### 8.1 成功发送短信
```bash
curl -X POST http://localhost:8088/v2/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "amount": "0.1",
    "content": "测试短信内容",
    "mobile": "13800138000",
    "sms_autograph": "102960",
    "app_name": "xiaolu"
  }'
```

#### 8.2 查询短信订单
```bash
curl -X GET "http://localhost:8088/v2/sms/orders?user_id=user123"
```

#### 8.3 测试短信上行回复
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

#### 8.4 查询短信回复
```bash
curl -X GET "http://localhost:8088/v2/sms/replies?user_id=user123"
```

### 9. 部署说明

1. 执行数据库迁移脚本
2. 更新 config.yaml 中的短信 API 配置
3. 重启服务
4. 验证短信发送功能

## 注意事项

1. 确保用户有足够余额才能发送短信
2. 短信签名 ID 需要提前配置
3. 发件人账号必须是验证过的真实号码
4. 定时发送时间使用秒级时间戳
5. API 调用失败会自动重试 3 次
6. 短信上行回调只支持 HTTP 协议，不要强制 HTTPS
7. 回调接口为公开接口，不需要认证
8. api_send_id 是关联短信发送和回复的关键字段
