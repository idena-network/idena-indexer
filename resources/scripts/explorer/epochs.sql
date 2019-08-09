select e.epoch,
       (select count(*)
        from epoch_identities ei
                 join address_states s on s.id = ei.address_state_id
        where ei.epoch_id = e.id
          and s.state in ('Verified', 'Newbie')) validated_count,
       (select count(*)
        from blocks b
        where b.epoch_id = e.id)                 block_count,
       (select count(*)
        from transactions t,
             blocks b
        where t.block_id = b.id
          and b.epoch_id = e.id)                 tx_count,
       (select count(*)
        from transactions t,
             blocks b
        where t.block_id = b.id
          and b.epoch_id = e.id
          and t.type = 'InviteTx')               invite_count,
       (select count(*)
        from flips f,
             transactions t,
             blocks b
        where f.tx_id = t.id
          and t.block_id = b.id
          and b.epoch_id = e.id)                 flip_count
from epochs e
order by e.epoch
limit $2
    offset $1