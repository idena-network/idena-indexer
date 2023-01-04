CREATE TABLE IF NOT EXISTS latest_activation_txs
(
    activation_tx_id bigint  NOT NULL,
    epoch            integer NOT NULL,
    address_id       bigint  NOT NULL
);

INSERT INTO latest_activation_txs (activation_tx_id, epoch, address_id)
SELECT act.tx_id, b.epoch, t.to
FROM activation_txs act
         JOIN transactions t ON t.id = act.tx_id
         JOIN blocks b ON b.height = t.block_height AND b.epoch >= (SELECT max(epoch) - 2 FROM epochs)