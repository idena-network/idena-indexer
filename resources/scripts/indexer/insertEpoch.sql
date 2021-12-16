insert into epochs (epoch, validation_time, root)
values ($1, $2, $3)
on conflict do nothing;
