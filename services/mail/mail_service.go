package mail

import (
	"phone_ios_backend/config"
	"phone_ios_backend/logger"
	"phone_ios_backend/my_utils"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"
)

// MailService 邮件服务
type MailService struct {
	config *config.Config
}

// NewMailService 创建邮件服务
func NewMailService() *MailService {
	return &MailService{
		config: config.Setting,
	}
}

// SendManualMessageOrderNotification 发送手动消息订单通知邮件
func (m *MailService) SendManualMessageOrderNotification(orderNo, userID, platform, platformAccount, content string) error {
	mailConfig := m.config.APP.Mail

	// 邮件主题
	subject := "人工传话平台订单通知"

	// 邮件正文
	body := fmt.Sprintf("订单号：%s\n用户ID：%s\n传话平台：%s\n传话账号：%s\n传话内容：%s", orderNo, userID, platform, platformAccount, content)

	// 每封邮件生成唯一 Message-ID，避免被中转/收件服务器去重丢弃
	messageID := fmt.Sprintf("<%d.%s@%s>", time.Now().UnixNano(), my_utils.GenerateRandomString(8), mailConfig.SMTPHost)
	date := time.Now().Format(time.RFC1123Z)

	// 构建邮件内容
	message := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMessage-ID: %s\r\nDate: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		mailConfig.FromName,
		mailConfig.Username,
		mailConfig.NotifyEmail,
		subject,
		messageID,
		date,
		body,
	)

	// SMTP 认证
	auth := smtp.PlainAuth("", mailConfig.Username, mailConfig.Password, mailConfig.SMTPHost)

	// SMTP 地址
	addr := fmt.Sprintf("%s:%d", mailConfig.SMTPHost, mailConfig.SMTPPort)

	// 发送邮件（使用 TLS）
	err := m.sendMailWithTLS(addr, auth, mailConfig.Username, []string{mailConfig.NotifyEmail}, []byte(message))
	if err != nil {
		logger.Logger.Errorf("Failed to send manual message order notification email: %v", err)
		return err
	}

	logger.Logger.Infof("Manual message order notification email sent successfully to: %s", mailConfig.NotifyEmail)
	return nil
}

// sendMailWithTLS 使用 TLS 发送邮件
func (m *MailService) sendMailWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// 分离地址的主机和端口
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	// 创建 TLS 连接
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	// 认证
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	// 设置发件人
	if err := client.Mail(from); err != nil {
		return err
	}

	// 设置收件人
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	// 获取 DataWriter
	writer, err := client.Data()
	if err != nil {
		return err
	}

	// 写入邮件内容并关闭 writer（发送 <CRLF>.<CRLF> 结束标记）
	if _, err = writer.Write(msg); err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}

	// 发送 QUIT 命令正常结束 SMTP 会话，否则服务端会保留连接占用并发配额
	return client.Quit()
}
