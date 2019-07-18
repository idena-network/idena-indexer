insert into address_states (address_id, state, is_actual)
values ((select id from addresses where address = $1), $2, true) returning id