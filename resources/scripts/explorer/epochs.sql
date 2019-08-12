select e.epoch,
       (select count(*)
        from epoch_identities ei
                 join address_states s on s.id = ei.address_state_id
        where ei.epoch = e.epoch
          and s.state in ('Verified', 'Newbie')) validated_count,
       (select count(*)
        from blocks b
        where b.epoch = e.epoch)                 block_count,
       (select count(*)
        from transactions t,
             blocks b
        where t.block_height = b.height
          and b.epoch = e.epoch)                 tx_count,
       (select count(*)
        from transactions t,
             blocks b
        where t.block_height = b.height
          and b.epoch = e.epoch
          and t.type = 'InviteTx')               invite_count,
       (select count(*)
        from flips f,
             transactions t,
             blocks b
        where f.tx_id = t.id
          and t.block_height = b.height
          and b.epoch = e.epoch)                 flip_count
from epochs e
order by e.epoch desc
limit $2
    offset $1