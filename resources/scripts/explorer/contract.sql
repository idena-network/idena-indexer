SELECT dict.name              "type",
       authora.address        author,
       deployt.hash           deployTxHash,
       deployb.timestamp      deployTxTimestamp,
       terminationt.hash      terminationTxHash,
       terminationb.timestamp terminationTxTimestamp
FROM contracts c
         JOIN dic_contract_types dict on dict.id = c.type
         JOIN addresses a ON a.id = c.contract_address_id AND lower(a.address) = lower($1)
         JOIN transactions deployt ON deployt.id = c.tx_id
         JOIN blocks deployb on deployb.height = deployt.block_height
         JOIN addresses authora ON authora.id = deployt.from

         LEFT JOIN time_lock_contract_terminations tlct ON c.type = 1 AND tlct.tl_contract_tx_id = c.tx_id
         LEFT JOIN oracle_voting_contract_terminations ovct ON c.type = 2 AND ovct.ov_contract_tx_id = c.tx_id
         LEFT JOIN oracle_lock_contract_terminations olct ON c.type = 3 AND olct.ol_contract_tx_id = c.tx_id
         LEFT JOIN multisig_contract_terminations mct ON c.type = 4 AND mct.ms_contract_tx_id = c.tx_id
         LEFT JOIN refundable_oracle_lock_contract_terminations rolct
                   ON c.type = 5 AND rolct.ol_contract_tx_id = c.tx_id

         LEFT JOIN transactions terminationt ON (c.type = 1 AND terminationt.id = tlct.termination_tx_id OR
                                                 c.type = 2 AND terminationt.id = ovct.termination_tx_id OR
                                                 c.type = 3 AND terminationt.id = olct.termination_tx_id OR
                                                 c.type = 4 AND terminationt.id = mct.termination_tx_id OR
                                                 c.type = 5 AND terminationt.id = rolct.termination_tx_id)
         LEFT JOIN blocks terminationb on terminationb.height = terminationt.block_height
