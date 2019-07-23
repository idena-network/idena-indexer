insert into address_states (address_id, state, is_actual, block_id)
values ($1, $2, true, $3) returning id