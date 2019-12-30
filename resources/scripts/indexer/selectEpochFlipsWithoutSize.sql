select cid
from flips
where size = 0
  and tx_id in (
    select id
    from transactions
    where block_height in
          (select height from blocks where epoch = $1)
)
limit $2;