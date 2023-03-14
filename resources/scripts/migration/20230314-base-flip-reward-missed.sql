do
$$
    declare
        l_record       record;
        l_flips_share  numeric;
        l_flips_missed numeric;
    begin
        create table epoch_103_backup_validation_reward_summaries
        (
            address_id          bigint,
            flips_missed        numeric(30, 18),
            flips_missed_reason smallint
        );

        insert into epoch_103_backup_validation_reward_summaries
        select address_id, flips_missed, flips_missed_reason
        from validation_reward_summaries
        where epoch = 103;

        select flips_share into l_flips_share from total_rewards where epoch = 103;

        for l_record in select vrs.*, ei.available_flips
                        from validation_reward_summaries vrs
                                 left join epoch_identities ei
                                           on ei.epoch = vrs.epoch and ei.address_id = vrs.address_id
                        where vrs.epoch = 103
                          and coalesce(vrs.extra_flips_missed, 0) > 0
            loop
                l_flips_missed = l_flips_share;
                if l_record.available_flips = 5 and l_record.extra_flips is null then
                    l_flips_missed = l_flips_missed + l_flips_share;
                end if;

                update validation_reward_summaries
                set flips_missed        = coalesce(l_record.flips_missed, 0) + l_flips_missed,
                    flips_missed_reason = coalesce(l_record.flips_missed_reason, 4)
                where epoch = l_record.epoch
                  and address_id = l_record.address_id;

            end loop;
    end;
$$;