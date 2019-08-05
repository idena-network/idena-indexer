select count(*) answer_count
from answers a
         join flips f on f.id = a.flip_id
where lower(f.cid) = lower($1)
  and a.is_short = $2