CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,                     -- Recipient User ID
    type VARCHAR(255) NOT NULL,
    entity_type VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    entity_id UUID NOT NULL,                   -- ID of the related entity
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    -- Optional: Add foreign keys for entity_id if they always refer to a specific table
    -- based on entity_type, though this is harder to enforce directly in SQL
    -- for a polymorphic association. Application-level integrity would be key.
);

-- Indexes for performance
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_user_id_is_read ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);