SELECT tr.raw
FROM transactions t
         JOIN transaction_raws tr on tr.tx_id = t.id
WHERE lower(t.hash) = lower($1)