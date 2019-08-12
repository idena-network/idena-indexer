insert into balances (address_id, balance, stake, block_height)
values ((select id from addresses where address = $1), $2, $3, $4)