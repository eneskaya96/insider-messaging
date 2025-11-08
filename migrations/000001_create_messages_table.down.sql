DROP INDEX IF EXISTS idx_messages_pending_fifo;
DROP INDEX IF EXISTS idx_messages_status_created_at;
DROP INDEX IF EXISTS idx_messages_sent_at;
DROP INDEX IF EXISTS idx_messages_created_at;
DROP INDEX IF EXISTS idx_messages_status;

DROP TABLE IF EXISTS messages;
