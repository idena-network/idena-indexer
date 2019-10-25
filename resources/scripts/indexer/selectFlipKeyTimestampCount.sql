select count(*)
from flip_key_timestamps
where lower(address) = lower($1)
  and epoch = $2