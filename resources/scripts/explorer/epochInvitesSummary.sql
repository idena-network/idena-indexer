select (select count(*)
        from transactions t
                 join blocks b on b.id = t.block_id
        where b.epoch_id = e.id
          and t.Type = 'InviteTx'
       ) all_count,
       0 used_count
from epochs e
where e.epoch = $1