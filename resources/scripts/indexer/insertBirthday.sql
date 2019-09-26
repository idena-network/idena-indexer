insert into birthdays (address_id, birth_epoch)
values ((select id from addresses where lower(address) = lower($1)), $2)