select fp.data
from flip_pics fp
where fp.fd_flip_tx_id = (select tx_id from flips where lower(cid) = lower($1))
order by fp.index