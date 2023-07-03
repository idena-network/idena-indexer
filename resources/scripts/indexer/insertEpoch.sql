insert into epochs (epoch, validation_time, root, discrimination_stake_threshold)
values ($1, $2, $3, $4)
on conflict do nothing;
