insert into penalties (address_id, penalty, penalty_seconds, block_height)
values ((select id from addresses where lower(address) = lower($1)), $2, $3, $4)