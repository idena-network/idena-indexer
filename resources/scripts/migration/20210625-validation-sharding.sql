ALTER TABLE epoch_identities
    ADD COLUMN shard_id integer;
ALTER TABLE epoch_identities
    ADD COLUMN new_shard_id integer;
ALTER TABLE activation_txs
    ADD COLUMN shard_id integer;

DROP TYPE tp_epoch_identity;
DROP TYPE tp_activation_tx;