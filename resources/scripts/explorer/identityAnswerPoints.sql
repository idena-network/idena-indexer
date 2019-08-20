select a.short_point,
       a.short_flips,
       a.total_short_point,
       a.total_short_flips,
       a.long_point,
       a.long_flips
from (select sum(ei.short_point)       short_point,
             sum(ei.short_flips)       short_flips,
             max(ei.total_short_point) total_short_point,
             max(ei.total_short_flips) total_short_flips,
             sum(ei.long_point)        long_point,
             sum(ei.long_flips)        long_flips
      from epoch_identities ei
               join address_states s on s.id = ei.address_state_id
               join addresses a on a.id = s.address_id
      where lower(a.address) = lower($1)
        and s.block_height >= (select coalesce(max(block_height), 0)
                               from address_states s
                               where s.address_id = a.id
                                 and s.state = 'Candidate')
      group by s.address_id) a