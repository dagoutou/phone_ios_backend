# Apple IAP 积分充值 API 文档

## 概览

iOS App 内销售数字商品(积分)按 App Store 审核规则必须使用 IAP,本服务接收 iOS 端上传的支付凭证(receipt),完成校验后发放积分。

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v2/payments/apple_iap/verify` | 上传 receipt 验单并发放积分 |
| GET  | `/v2/points/balance` | 查询当前积分余额 |
| GET  | `/v2/points/transactions` | 分页查询积分变动明细 |

基础地址:开发环境 `http://112.74.35.176:8089`,生产环境按实际域名。

> 这三个接口当前**不走 JWT 鉴权**(router 已配置 `payments` 和 `points` 路径跳过 token 校验)。`user_id` 由请求方传入,生产环境建议加上签名或改走 JWT。

---

## 1. 验单发放积分

### `POST /v2/payments/apple_iap/verify`

iOS 端通过 StoreKit 完成内购后,把苹果返回的 receipt-data(base64)上传到此接口,后端调用苹果验单服务器校验真伪,通过后发放对应积分。

### 请求体

```json
{
  "user_id": "u_123456",
  "app_name": "xiaolu",
  "receipt": "<base64 编码的支付凭证,从 StoreKit 拿到的 transactionReceipt>",
  "product_id": "cn.xiaolu.credits60"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `user_id` | string | 是 | 接收积分的用户 ID |
| `app_name` | string | 是 | 应用标识,目前固定为 `xiaolu` |
| `receipt` | string | 是 | base64 字符串,StoreKit 返回的 `payment.transactionReceipt` |
| `product_id` | string | 是 | App Store Connect 后台配置的内购商品 ID |

### curl 示例

```bash
curl -X POST 'http://112.74.35.176:8089/v2/payments/apple_iap/verify' \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "u_123456",
    "app_name": "xiaolu",
    "receipt": "MIIBOgIBAAJBAK3y...",
    "product_id": "cn.xiaolu.credits60"
  }'
```

### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "transaction_id": "1000000123456789",
    "product_id": "cn.xiaolu.credits60",
    "credits": 60,
    "balance": 360,
    "environment": "Sandbox"
  }
}
```

| 字段 | 说明 |
|------|------|
| `transaction_id` | 苹果返回的交易号,前端可用于幂等去重 |
| `product_id` | 内购商品 ID |
| `credits` | 本次发放的积分数 |
| `balance` | 用户当前积分余额(发放后) |
| `environment` | 验单使用的环境,`Sandbox` 或 `Production` |

### 错误响应

```json
{
  "code": 400,
  "message": "apple verify failed: status=21002"
}
```

常见苹果 status 码:

| status | 含义 | 处理建议 |
|--------|------|---------|
| 0 | 验证成功 | — |
| 21000 | 未使用 POST 请求 | 后端 bug |
| 21002 | receipt 数据格式错误 | 客户端检查 receipt 来源 |
| 21003 | receipt 无法通过身份验证 | 检查 sharedSecret 配置 |
| 21005 | receipt 当前不可用 | 苹果服务器问题,稍后重试 |
| 21007 | 沙盒收据发到了生产环境 | **后端自动回退沙盒,客户端无需处理** |
| 21008 | 生产收据发到了沙盒 | 后端自动回退生产 |

业务层错误:

| HTTP code | message | 说明 |
|-----------|---------|------|
| 400 | `apple iap shared secret not configured` | 后端配置未填,联系运维 |
| 400 | `no matching in_app transaction for product_id` | receipt 里没有此 product_id 的交易 |
| 400 | `product not configured: <product_id>` | product_prices 表没有此商品,需要 INSERT |
| 400 | `bundle_id mismatch` | receipt 的 bundle_id 和后端配置不一致 |
| 409 | `transaction already granted` | **同一 transaction_id 已发放过(防重放),前端可静默忽略** |
| 429 | `receipt is being processed` | 同一 receipt 并发提交被 redis 锁拦截,稍后重试 |

### iOS 客户端集成要点

```swift
// 伪代码:Swift / StoreKit 2
let result = try await product.purchase()
switch result {
case .success(let verification):
    let transaction = try checkVerified(verification)
    let receiptData = transaction.payloadJSON /* or legacy: payment.transactionReceipt */
    let receiptBase64 = receiptData.base64EncodedString()
    API.post("/v2/payments/apple_iap/verify", body: [
        "user_id": userId,
        "app_name": "xiaolu",
        "receipt": receiptBase64,
        "product_id": transaction.productID
    ]) { resp in
        if resp.code == 200 {
            // 展示积分到账,可调用 await transaction.finish() 完成订单
        } else if resp.message.contains("already granted") {
            // 已发放过,也调用 finish() 清除队列
        }
    }
}
```

**重要**:只在 verify 接口返回成功(或返回"已发放过")后才调用 `transaction.finish()`,否则保留在队列里下次启动重试。

---

## 2. 查询积分余额

### `GET /v2/points/balance`

### 请求参数(Query)

| 参数 | 必填 | 说明 |
|------|------|------|
| `user_id` | 是 | 用户 ID |
| `app_name` | 是 | 应用标识 |

### 示例

```bash
curl 'http://112.74.35.176:8089/v2/points/balance?user_id=u_123456&app_name=xiaolu'
```

### 响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "u_123456",
    "balance": 360
  }
}
```

> 用户从未充值过时,`balance` 返回 `0`,不会报错。

---

## 3. 查询积分明细

### `GET /v2/points/transactions`

### 请求参数(Query)

| 参数 | 必填 | 说明 |
|------|------|------|
| `user_id` | 是 | 用户 ID |
| `app_name` | 是 | 应用标识 |
| `page` | 否 | 页码,默认 1 |
| `size` | 否 | 每页条数,默认 20,最大 200 |

### 示例

```bash
curl 'http://112.74.35.176:8089/v2/points/transactions?user_id=u_123456&app_name=xiaolu&page=1&size=20'
```

### 响应

```json
{
  "code": 200,
  "message": "success",
  "count": 3,
  "data": {
    "list": [
      {
        "id": 102,
        "user_id": "u_123456",
        "app_name": "xiaolu",
        "type": "recharge",
        "amount": 60,
        "balance_after": 360,
        "source": "iap",
        "ref_order_no": "1000000123456789",
        "product_id": "cn.xiaolu.credits60",
        "remark": "Apple IAP 充值",
        "create_time": "2026-08-09T15:30:00Z"
      }
    ],
    "total": 3
  }
}
```

| 字段 | 说明 |
|------|------|
| `type` | `recharge`(充值) / `consume`(消费) / `refund`(退回) |
| `amount` | 变动数量,正数为增加 |
| `balance_after` | 本次变动后的余额 |
| `source` | 来源,目前只有 `iap`,后续可能扩展到 `alipay` / `wechat` / `manual` / `admin` |
| `ref_order_no` | 关联订单号,IAP 场景为苹果的 `original_transaction_id` |

---

## 上线检查清单

- [ ] 数据库已执行迁移:`006` / `007` / `008` / `009` / `010`(套餐 seed 数据)
- [ ] App Store Connect 后台已配置 6 个内购商品,product_id 与 `010_seed_apple_iap_products.sql` 一致
- [ ] App Store Connect 后台已生成 **App 专用共享密钥**,并填入 `config/config.yaml` 的 `AppleIAP.SharedSecret`
- [ ] `config/config.yaml` 的 `AppleIAP.BundleID` 与 iOS 工程的 Bundle Identifier 一致
- [ ] 沙盒账号测试通过:小额购买 → verify 返回 200 → 余额增加 → 再次 verify 返回 409
- [ ] TestFlight 包验证真实环境一次小额购买
