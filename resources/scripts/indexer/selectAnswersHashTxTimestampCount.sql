select count(*)
from answers_hash_tx_timestamps
where lower(address) = lower($1)
  and epoch = $2