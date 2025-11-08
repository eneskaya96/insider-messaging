CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP,
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    last_error TEXT,
    error_code VARCHAR(50),
    webhook_message_id VARCHAR(255),
    webhook_response TEXT,
    version BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT chk_status CHECK (status IN ('pending', 'processing', 'sent', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_sent_at ON messages(sent_at) WHERE sent_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_messages_status_created_at ON messages(status, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_pending_fifo ON messages(created_at) WHERE status = 'pending';

COMMENT ON TABLE messages IS 'Stores all messages to be sent via webhook';
COMMENT ON COLUMN messages.status IS 'Message status: pending, processing, sent, failed';
COMMENT ON COLUMN messages.attempts IS 'Number of send attempts made';
COMMENT ON COLUMN messages.max_attempts IS 'Maximum number of retry attempts allowed';
COMMENT ON COLUMN messages.version IS 'Version number for optimistic locking';
COMMENT ON COLUMN messages.webhook_message_id IS 'Message ID returned by webhook after successful send';
