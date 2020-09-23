ALTER TABLE answers
    ADD COLUMN grade smallint;
UPDATE answers
SET grade = (case when wrong_words then 1 else 2 end);
ALTER TABLE answers
    ALTER COLUMN grade SET NOT NULL;
ALTER TABLE answers
    DROP COLUMN wrong_words;


ALTER TABLE flips
    ADD COLUMN grade smallint;
UPDATE flips
SET grade = (case
                 when wrong_words is not null and wrong_words then 1
                 when wrong_words is not null and not wrong_words then 2
                 else null end);
ALTER TABLE flips
    DROP COLUMN wrong_words;

DROP TYPE tp_answer CASCADE;
DROP TYPE tp_flip_state CASCADE;