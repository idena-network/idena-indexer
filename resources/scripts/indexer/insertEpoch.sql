insert into epochs (epoch, validation_time)
values ($1, $2) RETURNING id