select t.Hash,
       dtt.name                                 "type",
       b.Timestamp,
       afrom.Address                            "from",
       COALESCE(ato.Address, '')                "to",
       t.Amount,
       t.Tips,
       t.max_fee,
       t.Fee,
       t.size,
       coalesce(atxs.balance_transfer,
                coalesce(ktxs.stake_transfer,
                         kitxs.stake_transfer)) transfer
from transactions t
         join blocks b on b.height = t.block_height
         join addresses afrom on afrom.id = t.from
         left join addresses ato on ato.id = t.to
         join dic_tx_types dtt on dtt.id = t.Type
         left join activation_tx_transfers atxs on atxs.tx_id = t.id and t.type = 1
         left join kill_tx_transfers ktxs on ktxs.tx_id = t.id and t.type = 3
         left join kill_invitee_tx_transfers kitxs on kitxs.tx_id = t.id and t.type = 10
where (select id from addresses where lower(address) = lower($1))
          in (t.from, t.to)
order by b.height desc
limit $3
offset
$2