SELECT tlc.timestamp
FROM contracts c
         JOIN time_lock_contracts tlc ON tlc.contract_tx_id = c.tx_id
WHERE c.contract_address_id = (SELECT id FROM addresses WHERE lower(address) = lower($1))