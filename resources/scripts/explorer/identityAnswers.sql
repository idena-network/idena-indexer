select a.short_point,
       a.short_flips,
       (case when short_flips != 0 then short_point / short_flips else 0 end) short_score,
       a.long_point,
       a.long_flips,
       (case when long_flips != 0 then long_point / long_flips else 0 end)    long_score
from (select sum(ei.short_point) short_point,
             sum(ei.short_flips) short_flips,
             sum(ei.long_point)  long_point,
             sum(ei.long_flips)  long_flips
      from epoch_identities ei
      where ei.identity_id = $1
      group by ei.identity_id) a