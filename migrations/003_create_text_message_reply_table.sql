-- Create text_message_reply table for SMS upstream replies
CREATE TABLE IF NOT EXISTS text_message_reply (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    api_send_id VARCHAR(50) NOT NULL COMMENT 'API send ID from original SMS',
    user_id VARCHAR(64) NOT NULL,
    order_no VARCHAR(64) NOT NULL,
    domain VARCHAR(100) COMMENT 'Authorized domain',
    content TEXT NOT NULL COMMENT 'Reply content',
    mobile VARCHAR(20) NOT NULL COMMENT 'Reply phone number',
    create_time DATETIME COMMENT 'Reply time from SMS system',
    lang VARCHAR(10) COMMENT 'System language',
    created_at DATETIME NOT NULL,
    INDEX idx_api_send_id (api_send_id),
    INDEX idx_user_id (user_id),
    INDEX idx_order_no (order_no),
    INDEX idx_mobile (mobile)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
