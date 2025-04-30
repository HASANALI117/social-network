-- +migrate Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE group_invitations (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    inviter_id TEXT NOT NULL,
    invitee_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, accepted, rejected
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (inviter_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (invitee_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (group_id, invitee_id) -- A user can only have one pending invitation per group
);

CREATE TABLE group_join_requests (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    requester_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, accepted, rejected
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (requester_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (group_id, requester_id) -- A user can only have one pending request per group
);

-- Trigger to update 'updated_at' timestamp on row update
CREATE TRIGGER update_group_invitations_updated_at
AFTER UPDATE ON group_invitations
FOR EACH ROW
BEGIN
    UPDATE group_invitations SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

CREATE TRIGGER update_group_join_requests_updated_at
AFTER UPDATE ON group_join_requests
FOR EACH ROW
BEGIN
    UPDATE group_join_requests SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Indexes for faster lookups
CREATE INDEX idx_group_invitations_group_invitee ON group_invitations (group_id, invitee_id);
CREATE INDEX idx_group_invitations_invitee_status ON group_invitations (invitee_id, status);

CREATE INDEX idx_group_join_requests_group_requester ON group_join_requests (group_id, requester_id);
CREATE INDEX idx_group_join_requests_group_status ON group_join_requests (group_id, status);

-- +migrate StatementEnd
