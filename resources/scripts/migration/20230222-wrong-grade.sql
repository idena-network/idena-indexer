DROP TYPE tp_epoch_identity CASCADE;

ALTER TABLE epoch_identities
    ADD COLUMN wrong_grade_reason smallint;