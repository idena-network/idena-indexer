select sum(1) all_count, sum((case when invite_tx_id is not null then 1 else 0 end)) used_count
from transactions t
         join blocks b
              on b.id = t.block_id
         join epochs e on e.id = b.epoch_id
         left join used_invites ui on ui.invite_tx_id = t.id
where e.epoch = $1
  and t.type = 'InviteTx'