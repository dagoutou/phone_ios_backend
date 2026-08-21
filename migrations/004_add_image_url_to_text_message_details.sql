-- Add image_url and platform columns to text_message_details
ALTER TABLE text_message_details
    ADD COLUMN platform VARCHAR(50) DEFAULT '' COMMENT 'Manual message platform' AFTER content,
    ADD COLUMN image_url VARCHAR(255) DEFAULT '' COMMENT 'Completion image URL' AFTER type;
