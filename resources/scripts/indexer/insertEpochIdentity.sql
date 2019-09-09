insert into epoch_identities (epoch, address_state_id, short_point, short_flips, total_short_point,
                              total_short_flips, long_point, long_flips, approved, missed,
                              required_flips, made_flips)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id