SELECT a.address
FROM sorted_oracle_voting_contracts sovc
         LEFT JOIN contracts c ON tx_id = sovc.contract_tx_id
         LEFT JOIN addresses a ON a.id = c.contract_address_id
WHERE sovc.state in (1, 3)