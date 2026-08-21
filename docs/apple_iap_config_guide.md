# Apple IAP PrivateKey 配置指南

## 配置说明

在 `config/config.yaml` 中配置 Apple IAP 私钥时，需要注意 YAML 块标量 `|` 的缩进规则。

## 正确配置示例

```yaml
AppleIAP:
  BundleID: "com.lovephone.app"
  Environment: "sandbox"
  IssuerID: "69c49078-20d6-4e81-8e26-62bf203768b2"
  KeyID: "AAK6HTZC2B"
  PrivateKey: |                              # .p8 文件完整内容(含 BEGIN/END 行)
    -----BEGIN PRIVATE KEY-----
    MIGTAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBHkwdwIBAQQgAo1aTfLSCo7peKfG
    JVZf7RoMlW/8Dg/pcfpauje+qFKgCgYIKoZIzj0DAQehRANCAARh4p/7PLv4bVmB
    7ZG6NMTU+T4/HUN2NySHJlMpMhRqpxm6Gm4/IQjuJSuy9EMvMbGs0nSR55kGBD5Q
    AbBQo+HB
    -----END PRIVATE KEY-----
```

## 关键要点

### 1. 缩进必须一致 ⚠️
使用 `|` 块标量时，**所有行必须保持相同的缩进**：

```yaml
# ✅ 正确 - 所有行都是 4 个空格缩进（相对 PrivateKey）
  PrivateKey: |
    -----BEGIN PRIVATE KEY-----    # ← 4 个空格
    MIGTAgEAMBMGByqGSM49AgEG...    # ← 4 个空格（必须一致）
    -----END PRIVATE KEY-----       # ← 4 个空格

# ❌ 错误 - 缩进不一致，会导致 YAML 解析失败
  PrivateKey: |
    -----BEGIN PRIVATE KEY-----      # ← 4 个空格
  MIGTAgEAMBMGByqGSM49AgEGCCqGSM49  # ← 2 个空格（错误！）
```

### 2. 保留 BEGIN/END 行
私钥内容必须包含完整的 PEM 格式边界行：

```
-----BEGIN PRIVATE KEY-----
<base64编码的密钥内容>
-----END PRIVATE KEY-----
```

### 3. 如何获取 .p8 文件

1. 登录 [App Store Connect](https://appstoreconnect.apple.com/)
2. 进入 **用户和访问** > **集成** > **App 管理**
3. 点击 **密钥** 标签
4. 点击 **创建密钥** 或选择现有密钥
5. 下载 `.p8` 文件（**只能下载一次**，请妥善保管）

### 4. 密钥参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| BundleID | iOS App Bundle Identifier | `com.lovephone.app` |
| IssuerID | 在密钥页面顶部显示 | `69c49078-20d6-4e81-8e26-62bf203768b2` |
| KeyID | 创建密钥后显示 | `AAK6HTZC2B` |
| Environment | 运行环境 | `sandbox`(测试) 或 `production`(生产) |
| PrivateKey | .p8 文件完整内容 | 包含 BEGIN/END 行的完整 PEM 格式 |

## 验证配置

### 方法1：使用 Ruby 验证 YAML 解析
```bash
ruby -ryaml -e 'key = YAML.load_file("config/config.yaml").dig("APP","AppleIAP","PrivateKey"); puts key.include?("BEGIN PRIVATE KEY")'
```

### 方法2：查看私钥长度
正常解析后私钥内容长度约为 250-300 字节。

### 方法3：检查配置加载
```bash
go run main.go  # 启动服务，检查是否有配置相关错误
```

## 安全提示

### ⚠️ 私钥安全

1. **不要将 .p8 文件提交到版本控制系统**
   - 在 `.gitignore` 中添加：`config/config.yaml` 或使用环境变量

2. **不同环境使用不同密钥**
   - 测试环境：使用 Sandbox 密钥
   - 生产环境：使用 Production 密钥

3. **定期轮换密钥**
   - Apple 建议定期更新 API 密钥
   - 旧密钥过期前创建新密钥

4. **限制密钥权限**
   - 在 App Store Connect 中为每个密钥分配最小必要权限

## 常见错误

### YAML 解析错误
```
Error: could not find expected ':' at line 58
```
**原因**：PrivateKey 块的缩进不一致  
**解决**：确保所有行缩进相同（通常是 4 个空格）

### 私钥格式错误
```
Error: invalid PEM block
```
**原因**：缺少 BEGIN/END 行或内容被截断  
**解决**：复制 .p8 文件的完整内容，包括边界行

### JWT 生成失败
```
Error: generate jwt: parse private key
```
**原因**：私钥内容错误或格式不正确  
**解决**：重新下载 .p8 文件并完整复制

## 环境配置

### 开发/测试环境
```yaml
Environment: "sandbox"
```
使用沙盒环境进行测试，交易不会实际扣费。

### 生产环境
```yaml
Environment: "production"
```
切换到生产环境前确保：
1. 所有测试已完成
2. 使用正式的 BundleID
3. 配置正确的 IssuerID 和 KeyID

## 相关文件

- 配置文件：`config/config.yaml`
- 配置结构：`config/apple_iap_config.go`
- API 客户端：`services/pay/apple_server_api.go`
- 验证服务：`services/pay/apple_iap.go`
