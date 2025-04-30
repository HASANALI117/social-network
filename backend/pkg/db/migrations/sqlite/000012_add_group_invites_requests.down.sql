-- +migrate Down
-- SQL in this section is executed when the migration is rolled back.

DROP TRIGGER IF EXISTS update_group_join_requests_updated_at;
DROP TRIGGER IF EXISTS update_group_invitations_updated_at;

DROP INDEX IF EXISTS idx_group_join_requests_group_status;
DROP INDEX IF EXISTS idx_group_join_requests_group_requester;

DROP INDEX IF EXISTS idx_group_invitations_invitee_status;
DROP INDEX IF EXISTS idx_group_invitations_group_invitee;

DROP TABLE IF EXISTS group_join_requests;
DROP TABLE IF EXISTS group_invitations;

-- +migrate StatementEnd
