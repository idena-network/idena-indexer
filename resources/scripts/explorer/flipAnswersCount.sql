select count(*) answer_count
from answers a
         join flips f on f.tx_id = a.flip_tx_id and lower(f.cid) = lower($1)
where a.is_short = $2