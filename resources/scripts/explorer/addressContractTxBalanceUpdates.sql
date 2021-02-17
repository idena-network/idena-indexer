SELECT bu.tx_id,
       t.hash                                                          txHash,
       dt.name                                                         txType,
       b.timestamp,
       afrom.address                                                   "from",
       coalesce(ato.address, '')                                       "to",
       (case when coalesce(tr.success, true) then t.amount else 0 end) amount,
       t.tips,
       t.max_fee,
       t.fee,
       abu.address                                                     address,
       ac.address                                                      contract_address,
       dct.name                                                        contract_type,
       bu.call_method,
       bu.balance_old,
       bu.balance_new,
       tr.success                                                      tx_receipt_success,
       tr.gas_used                                                     tx_receipt_gas_used,
       tr.gas_cost                                                     tx_receipt_gas_cost,
       tr.method                                                       tx_receipt_method,
       tr.error_msg                                                    tx_receipt_error_msg
FROM contract_tx_balance_updates bu
         JOIN transactions t on t.id = bu.tx_id
         JOIN blocks b on b.height = t.block_height
         JOIN dic_tx_types dt on dt.id = t.type
         JOIN addresses afrom on afrom.id = t.from
         LEFT JOIN addresses ato on ato.id = t.to
         JOIN addresses abu on abu.id = bu.address_id
         JOIN contracts c on c.tx_id = bu.contract_tx_id
         JOIN addresses ac on ac.id = c.contract_address_id
         JOIN dic_contract_types dct on dct.id = c.type
         LEFT JOIN tx_receipts tr on t.type in (15, 16, 17) and tr.tx_id = t.id
WHERE ($4::bigint IS NULL OR bu.tx_id <= $4)
  AND bu.address_id = (SELECT id FROM addresses WHERE lower(address) = lower($1))
  AND bu.contract_tx_id = (SELECT tx_id
                           FROM contracts
                           WHERE contract_address_id = (SELECT id FROM addresses WHERE lower(address) = lower($2)))
ORDER BY tx_id DESC
LIMIT $3