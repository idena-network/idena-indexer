SELECT dict.name "type"
FROM contracts c
         JOIN addresses a ON a.id = c.contract_address_id AND lower(a.address) = lower($1)
         JOIN dic_contract_types dict on dict.id = c.type