-- Create text_message_details table for SMS send details
CREATE TABLE IF NOT EXISTS text_message_details (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    api_send_id VARCHAR(50) COMMENT 'API send ID from SMS service',
    user_id VARCHAR(64) NOT NULL,
    order_no VARCHAR(64) UNIQUE NOT NULL,
    mobile VARCHAR(20) NOT NULL COMMENT 'Recipient phone number',
    account VARCHAR(20) COMMENT 'Sender phone number',
    content TEXT COMMENT 'SMS content',
    plan_time DATETIME COMMENT 'Scheduled send time',
    create_time DATETIME COMMENT 'API submit time',
    strip INT COMMENT 'Number of SMS segments',
    remarks VARCHAR(50) COMMENT 'Send status remarks',
    domain VARCHAR(100) COMMENT 'Authorized domain',
    created_at DATETIME NOT NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_order_no (order_no),
    INDEX idx_api_send_id (api_send_id),
    INDEX idx_mobile (mobile)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
