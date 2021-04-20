DROP TYPE tp_rewarded_invitation CASCADE;

ALTER TABLE rewarded_invitations
    ADD COLUMN epoch_height integer;