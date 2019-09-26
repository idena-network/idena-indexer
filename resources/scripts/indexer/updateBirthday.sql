update birthdays
set birth_epoch=$2
where address_id = (select id from addresses where lower(address) = lower($1))
returning address_id