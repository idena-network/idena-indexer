select t.Hash                    invite_hash,
       a.address                 invite_author,
       b.timestamp               invite_timestamp,
       coalesce(at.Hash, '')     activation_hash,
       coalesce(aa.address, '')  activation_author,
       coalesce(ab.timestamp, 0) activation_timestamp
from transactions t
         join addresses a on a.id = t.from
         join blocks b on b.id = t.block_id
         join epochs e on e.id = epoch_id
         left join used_invites ui on ui.invite_tx_id = t.id
         left join transactions at on at.id = ui.activation_tx_id
         left join blocks ab on ab.id = at.block_id
         left join addresses aa on aa.id = at.to
where t.type = 'InviteTx'
  and lower(a.address) = lower($1)
order by b.height
limit $3
    offset $2