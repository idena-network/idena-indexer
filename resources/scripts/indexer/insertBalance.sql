insert into balances (address_id, balance, stake)
values ((select id from addresses where lower(address) = lower($1)), $2, $3)