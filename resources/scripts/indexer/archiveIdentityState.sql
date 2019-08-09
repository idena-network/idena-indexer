update address_states
set is_actual= false
where address_id = (select id from addresses where address = $1)
  and is_actual returning id