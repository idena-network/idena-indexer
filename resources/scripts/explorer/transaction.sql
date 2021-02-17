select b.epoch,
       b.height,
       b.hash,
       t.Hash,
       dtt.name                                                                                     "type",
       b.Timestamp,
       afrom.Address                                                                                "from",
       COALESCE(ato.Address, '')                                                                    "to",
       t.Amount,
       t.Tips,
       t.max_fee,
       t.Fee,
       t.size,
       coalesce(atxs.balance_transfer,
                coalesce(ktxs.stake_transfer,
                         kitxs.stake_transfer))                                                     transfer,
       (case when online.tx_id is not null then true when offline.tx_id is not null then false end) become_online,
       tr.success                                                                                   tx_receipt_success,
       tr.gas_used                                                                                  tx_receipt_gas_used,
       tr.gas_cost                                                                                  tx_receipt_gas_cost,
       tr.method                                                                                    tx_receipt_method,
       tr.error_msg                                                                                 tx_receipt_error_msg
from transactions t
         join blocks b on b.height = t.block_height
         join addresses afrom on afrom.id = t.from
         left join addresses ato on ato.id = t.to
         join dic_tx_types dtt on dtt.id = t.Type
         left join activation_tx_transfers atxs on atxs.tx_id = t.id and t.type = 1
         left join kill_tx_transfers ktxs on ktxs.tx_id = t.id and t.type = 3
         left join kill_invitee_tx_transfers kitxs on kitxs.tx_id = t.id and t.type = 10
         left join become_online_txs online on online.tx_id = t.id and t.type = 9
         left join become_offline_txs offline on offline.tx_id = t.id and t.type = 9
         LEFT JOIN tx_receipts tr on t.type in (15, 16, 17) and tr.tx_id = t.id
where lower(t.Hash) = lower($1)