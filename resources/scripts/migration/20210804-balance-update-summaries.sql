DO
$$
    BEGIN
        CREATE TABLE IF NOT EXISTS balance_update_summaries
        (
            address_id  bigint          NOT NULL,
            balance_in  numeric(30, 18) NOT NULL,
            balance_out numeric(30, 18) NOT NULL,
            stake_in    numeric(30, 18) NOT NULL,
            stake_out   numeric(30, 18) NOT NULL,
            penalty_in  numeric(30, 18) NOT NULL,
            penalty_out numeric(30, 18) NOT NULL,
            CONSTRAINT balance_update_summaries_pkey PRIMARY KEY (address_id)
        );

        INSERT INTO balance_update_summaries
        SELECT address_id,
               sum(case when balance_new > balance_old then balance_new - balance_old else 0 end) balance_id,
               sum(case when balance_new < balance_old then balance_old - balance_new else 0 end) balance_out,
               sum(case when stake_new > stake_old then stake_new - stake_old else 0 end)         stake_id,
               sum(case when stake_new < stake_old then stake_old - stake_new else 0 end)         stake_out,
               sum(case
                       when coalesce(penalty_new, 0) > coalesce(penalty_old, 0)
                           then coalesce(penalty_new, 0) - coalesce(penalty_old, 0)
                       else 0 end)                                                                penalty_id,
               sum(case
                       when coalesce(penalty_new, 0) < coalesce(penalty_old, 0)
                           then coalesce(penalty_old, 0) - coalesce(penalty_new, 0)
                       else 0 end)                                                                penalty_out
        FROM balance_updates
        GROUP BY address_id;
    END
$$