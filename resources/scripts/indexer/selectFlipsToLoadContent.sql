select fq.cid, fq.key, fq.attempts, fpk.key private_key
from flips_queue fq
         join flips f on lower(f.cid) = lower(fq.cid)
         left join flip_private_keys fpk on fpk.flip_tx_id = f.tx_id
where fq.next_attempt_timestamp < $1
order by fq.next_attempt_timestamp
limit $2