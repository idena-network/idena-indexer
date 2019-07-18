select s.state, count(*)
from address_states s
where s.is_actual
group by s.state