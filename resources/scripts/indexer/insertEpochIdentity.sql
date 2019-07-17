insert into epoch_identities (epoch_id, identity_id, state, short_point, short_flips, long_point, long_flips, approved,
                              missed)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id