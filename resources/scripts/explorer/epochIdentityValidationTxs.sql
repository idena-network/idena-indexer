select t.Hash, t.Type, b.Timestamp, afrom.Address "from", COALESCE(ato.Address, '') "to", t.Amount, t.Fee, t.size
from transactions t
         join blocks b on b.height = t.block_height
         join addresses afrom on afrom.id = t.from
         left join addresses ato on ato.id = t.to
where b.epoch = $1
  and lower(afrom.address) = lower($2)
  and t.Type in ('SubmitAnswersHashTx', 'SubmitShortAnswersTx', 'SubmitLongAnswersTx', 'EvidenceTx')
order by b.height desc