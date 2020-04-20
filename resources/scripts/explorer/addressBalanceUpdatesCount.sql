select count(*)
from balance_updates
where address_id = (select id from addresses where lower(address) = lower($1))