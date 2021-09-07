ALTER TABLE epoch_identities
    ADD COLUMN shard_id integer;
ALTER TABLE epoch_identities
    ADD COLUMN new_shard_id integer;
ALTER TABLE activation_txs
    ADD COLUMN shard_id integer;

DROP TYPE tp_epoch_identity CASCADE;
DROP TYPE tp_activation_tx CASCADE;

DROP TYPE tp_total_epoch_reward CASCADE;
ALTER TABLE total_rewards
    ADD COLUMN reports numeric(30, 18);
ALTER TABLE total_rewards
    ADD COLUMN reports_share numeric(30, 18);

UPDATE pg_attribute
SET atttypmod = 200 + 4
WHERE attrelid = 'words_dictionary'::regclass
  AND attname = 'description';

UPDATE pg_attribute
SET atttypmod = 50 + 4
WHERE attrelid = 'words_dictionary'::regclass
  AND attname = 'name';
