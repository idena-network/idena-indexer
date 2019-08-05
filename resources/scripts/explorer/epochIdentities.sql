select a.address,
       s.state,
       (
           select as2.state
           from address_states as2
                    join blocks b on b.id = as2.block_Id
           where as2.address_id = s.address_id
             and b.height < (select height from blocks where id = s.block_id)
           order by b.height desc
           limit 1
       ) prev_state,
       ei.approved,
       ei.missed,
       ei.short_point,
       ei.short_flips,
       ei.long_point,
       ei.long_flips
from epoch_identities ei
         join epochs e on e.id = ei.epoch_id
         join address_states s on s.id = ei.address_state_id
         join addresses a on a.id = s.address_id
where e.epoch = $1
limit $3
    offset $2