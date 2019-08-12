insert into flips_data (flip_id, block_height, tx_id)
values ((select id from flips where lower(cid) = lower($1)), $2, $3) returning id