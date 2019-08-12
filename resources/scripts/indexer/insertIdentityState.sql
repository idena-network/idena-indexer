insert into address_states (address_id, state, is_actual, block_height, prev_id)
values ((select id from addresses where address = $1), $2, true, $3, $4) returning id