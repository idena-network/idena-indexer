select s.state, e.epoch, b.height, b.hash, coalesce(t.hash, '') tx_hash, b.timestamp
from address_states s
         join addresses a on a.id = s.address_id
         join blocks b on b.id = s.block_id
         join epochs e on e.id = b.epoch_id
         left join transactions t on t.id = s.tx_id
where lower(a.address) = lower($1)
order by b.height
limit $3
    offset $2