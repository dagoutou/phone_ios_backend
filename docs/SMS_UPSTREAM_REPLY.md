# 短信上行回复接收文档

## 概述

本文档描述短信上行回复的接收流程和实现。当用户回复短信时，短信系统会通过回调接口通知应用服务器。

## 回调配置

### 回调接口
- **接口地址：** `POST /v2/pull_sms_upstream`
- **协议：** HTTP（不支持 HTTPS 强制跳转）
- **数据类型：** `application/json`
- **认证：** 不需要（公开接口）

### 回调请求参数

```json
{
  "domain": "yuming.top",           // 授权域名
  "content": "测试回复",             // 上行回复内容
  "mobile": "18888888888",          // 回复的手机号码
  "api_send_id": "2208050956474807877",  // 发送的唯一ID
  "create_time": "2022-08-04 17:34:51",  // 上行回复时间
  "lang": "cn"                      // 系统语言
}
```

## 处理流程

```
短信系统
  ↓
POST /v2/pull_sms_upstream
  ↓
验证 api_send_id
  ↓
查询原短信记录 (text_message_details)
  ↓
创建回复记录 (text_message_reply)
  ↓
更新原记录备注 (remarks = "success")
  ↓
返回成功响应
```

## 短信状态推送

### 接收发送失败通知

**接口：** `POST /v2/pull_sms_status`

当短信发送失败时，短信系统会调用此接口通知应用服务器。

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
- 根据 `api_send_id` 更新 `text_message_details` 表的 `remarks` 为 "failed"

**示例：**
```bash
curl -X POST http://your-domain.com/v2/pull_sms_status \
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

## 数据库表

### text_message_reply 表

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

## API 接口

### 1. 接收上行回复（回调接口）

**接口：** `POST /v2/pull_sms_upstream`

**示例：**
```bash
curl -X POST http://your-domain.com/v2/pull_sms_upstream \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "yuming.top",
    "content": "用户回复的内容",
    "mobile": "18888888888",
    "api_send_id": "2208050956474807877",
    "create_time": "2022-08-04 17:34:51",
    "lang": "cn"
  }'
```

### 2. 查询用户的短信回复

**接口：** `GET /v2/sms/replies`

**参数：** `user_id`（用户ID）

**示例：**
```bash
curl -X GET "http://your-domain.com/v2/sms/replies?user_id=user123"
```

### 3. 查询订单的短信回复

**接口：** `GET /v2/sms/replies/by_order`

**参数：** `order_no`（订单号）

**示例：**
```bash
curl -X GET "http://your-domain.com/v2/sms/replies/by_order?order_no=Balance20240101120000123456"
```

## 完整流程示例

### 步骤 1：用户发送短信
```bash
curl -X POST http://your-domain.com/v2/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "amount": "0.1",
    "content": "您的验证码是123456",
    "mobile": "13800138000",
    "sms_autograph": "102960",
    "app_name": "xiaolu"
  }'
```

**响应：**
```json
{
  "code": 200,
  "data": {
    "status": "success",
    "order_no": "Balance20240101120000123456"
  }
}
```

### 步骤 2：短信系统发送短信
- 短信系统调用第三方 API 发送短信
- 获得 `api_send_id`：`2208050956474807877`
- 存储到 `text_message_details` 表

### 步骤 3：用户回复短信
用户回复短信内容："收到"

### 步骤 4：短信系统回调
```bash
curl -X POST http://your-domain.com/v2/pull_sms_upstream \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "yuming.top",
    "content": "收到",
    "mobile": "13800138000",
    "api_send_id": "2208050956474807877",
    "create_time": "2024-01-01 14:30:00",
    "lang": "cn"
  }'
```

### 步骤 5：系统处理
1. 根据 `api_send_id` 查询原短信记录
2. 在 `text_message_reply` 表创建回复记录
3. 更新 `text_message_details` 表的 `remarks` 为 "success"

### 步骤 6：应用查询回复
```bash
curl -X GET "http://your-domain.com/v2/sms/replies?user_id=user123"
```

## 数据状态变化

### text_message_details 表状态

| 阶段 | remarks | 说明 |
|------|---------|------|
| 发送成功 | "已提交" | 短信已提交到运营商 |
| 收到回复 | "success" | 用户已回复短信 |

### text_message_reply 表记录

每次收到用户回复时，创建新记录：
- `api_send_id`：关联原短信
- `content`：回复内容
- `mobile`：回复手机号
- `create_time`：回复时间

## 错误处理

| 错误 | 说明 |
|------|------|
| api_send_id 不存在 | 返回成功，不创建记录 |
| 参数格式错误 | 返回 400 错误 |
| 数据库操作失败 | 返回 500 错误 |

## 注意事项

1. **协议限制：** 只支持 HTTP 协议，不要强制 HTTPS
2. **重复回调：** 根据 `api_send_id` 可以避免重复处理
3. **时间格式：** 使用 "YYYY-MM-DD HH:MM:SS" 格式
4. **编码格式：** UTF-8
5. **字符转义：** JSON 中的中文需要正确转义

## 测试用例

### 测试用例 1：正常回复
```bash
curl -X POST http://localhost:8088/v2/pull_sms_upstream \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "yuming.top",
    "content": "测试",
    "mobile": "18888888888",
    "api_send_id": "2208050956474807877",
    "create_time": "2022-08-04 17:34:51",
    "lang": "cn"
  }'
```

### 测试用例 2：缺少必填参数
```bash
curl -X POST http://localhost:8088/v2/pull_sms_upstream \
  -H "Content-Type: application/json" \
  -d '{
    "content": "测试",
    "mobile": "18888888888"
  }'
```
预期：返回 400 错误

### 测试用例 3：不存在的 api_send_id
```bash
curl -X POST http://localhost:8088/v2/pull_sms_upstream \
  -H "Content-Type: application/json" \
  -d '{
    "api_send_id": "9999999999999999999",
    "content": "测试",
    "mobile": "18888888888",
    "create_time": "2022-08-04 17:34:51"
  }'
```
预期：返回成功，但不创建记录

### 测试用例 4：短信发送失败通知
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
预期：返回成功，更新 remarks 为 "failed"

### 测试用例 5：短信发送失败通知（缺少参数）
```bash
curl -X POST http://localhost:8088/v2/pull_sms_status \
  -H "Content-Type: application/json" \
  -d '{
    "remarks": "驳回",
    "account": "oxmYc5jaVCj4Vmc9p1AHG-NmUCQQ"
  }'
```
预期：返回 400 错误
