select e.validation_time,
       (select count(*)
        from transactions t
                 join blocks b on b.id = t.block_id
        where b.epoch_id = e.id
          and t.type = 'InviteTx'
       ) invites,
       (select count(*)
        from flips f
                 join transactions t on t.id = f.tx_id
                 join blocks b on b.id = t.block_id
        where b.epoch_id = e.id
       ) flips
from epochs e
where e.epoch = $1