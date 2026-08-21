-- Add reject_reason column to text_message_details for storing audit rejection reason
ALTER TABLE text_message_details
    ADD COLUMN reject_reason VARCHAR(500) DEFAULT '' COMMENT 'Audit rejection reason' AFTER image_url;
