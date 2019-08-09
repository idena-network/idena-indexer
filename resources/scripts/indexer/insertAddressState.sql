insert into address_states (address_id, state, is_actual, block_id, tx_id, prev_id)
values ($1, $2, true, $3, (select id from transactions where lower(hash) = lower($4)), $5) returning id