update balances
set balance=$2,
    stake=$3
where address_id = (select id from addresses where lower(address) = lower($1))
returning address_id