update address_states
set is_actual= false
where address_id = (select id from addresses where lower(address) = lower($1))
  and is_actual returning id