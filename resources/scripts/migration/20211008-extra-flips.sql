ALTER TABLE answers
    ADD COLUMN index smallint;

ALTER TABLE answers
    ADD COLUMN considered boolean;

ALTER TABLE flips_to_solve
    ADD COLUMN index smallint;

DROP TYPE tp_flip_to_solve CASCADE;
DROP TYPE tp_answer CASCADE;