DROP TYPE tp_flip_state CASCADE;

ALTER TABLE flips
    ADD COLUMN grade_score real;
