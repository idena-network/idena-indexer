update address_states
set is_actual= false
where address_id = $1
  and is_actual;