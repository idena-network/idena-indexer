insert into fund_rewards (address_id, block_height, balance, type)
values ((select id from addresses where lower(address)=lower($1)), $2, $3, $4)