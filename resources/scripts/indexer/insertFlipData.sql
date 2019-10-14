insert into flips_data (flip_id, block_height, tx_id)
values ($1, $2, $3)
returning id