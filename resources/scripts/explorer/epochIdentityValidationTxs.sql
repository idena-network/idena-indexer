select t.Hash,
       dtt.name                  "type",
       b.Timestamp,
       afrom.Address             "from",
       COALESCE(ato.Address, '') "to",
       t.Amount,
       t.Tips,
       t.max_fee,
       t.Fee,
       t.size,
       null                      transfer,
       false                     become_online
from transactions t
         join blocks b on b.height = t.block_height and b.epoch = $1
         join addresses afrom on afrom.id = t.from and lower(afrom.address) = lower($2)
         left join addresses ato on ato.id = t.to
         join dic_tx_types dtt on dtt.id = t.Type
where t.Type in (select id
                 from dic_tx_types
                 where name in ('SubmitAnswersHashTx', 'SubmitShortAnswersTx', 'SubmitLongAnswersTx', 'EvidenceTx'))
order by t.id desc