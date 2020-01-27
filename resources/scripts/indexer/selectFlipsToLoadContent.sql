select fq.cid, fq.key, fq.attempts
from flips_queue fq
where fq.next_attempt_timestamp < $1
order by fq.next_attempt_timestamp
limit $2