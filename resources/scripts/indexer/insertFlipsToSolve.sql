insert into flips_to_solve (epoch_identity_id, flip_id, is_short)
values ($1, $2, $3) RETURNING id