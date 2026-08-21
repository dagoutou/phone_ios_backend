-- Create orders table for text message orders
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_no VARCHAR(64) UNIQUE NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    order_type VARCHAR(20) NOT NULL COMMENT 'text_message for SMS orders',
    status VARCHAR(20) NOT NULL DEFAULT 'paid' COMMENT 'paid, pending, failed, cancelled',
    plan_time BIGINT COMMENT 'Scheduled send time (timestamp)',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_order_no (order_no),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
