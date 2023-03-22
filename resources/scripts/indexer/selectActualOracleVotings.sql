SELECT a.address
FROM sorted_oracle_voting_contracts sovc
         LEFT JOIN contracts c ON c.tx_id = sovc.contract_tx_id AND c.type = 2
         LEFT JOIN addresses a ON a.id = c.contract_address_id
WHERE sovc.state in (1, 3)