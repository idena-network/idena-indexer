SELECT dict.name "type", authora.address author
FROM contracts c
         JOIN addresses a ON a.id = c.contract_address_id AND lower(a.address) = lower($1)
         JOIN transactions t ON t.id = c.tx_id
         JOIN addresses authora ON authora.id = t.from
         JOIN dic_contract_types dict on dict.id = c.type