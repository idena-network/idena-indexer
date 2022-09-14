CREATE OR REPLACE PROCEDURE reset_changes_to(p_block_height bigint)
    LANGUAGE 'plpgsql'
AS
$$
DECLARE
    CHANGE_TYPE_ORACLE_VOTING_RESULTS           CONSTANT smallint = 1;
    CHANGE_TYPE_ORACLE_VOTING_SUMMARIES         CONSTANT smallint = 2;
    CHANGE_TYPE_SORTED_ORACLE_VOTINGS           CONSTANT smallint = 3;
    CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEES CONSTANT smallint = 4;
    CHANGE_TYPE_BALANCE_UPDATE_SUMMARIES        CONSTANT smallint = 5;
    CHANGE_TYPE_MINING_REWARD_SUMMARIES         CONSTANT smallint = 6;
    l_rec                                                record;
BEGIN
    for l_rec in SELECT id, type FROM changes WHERE block_height > p_block_height ORDER BY id DESC
        loop
            if l_rec.type = CHANGE_TYPE_ORACLE_VOTING_RESULTS then
                call reset_oracle_voting_contract_results_changes(l_rec.id);
                continue;
            end if;

            if l_rec.type = CHANGE_TYPE_ORACLE_VOTING_SUMMARIES then
                call reset_oracle_voting_contract_summaries_changes(l_rec.id);
                continue;
            end if;

            if l_rec.type = CHANGE_TYPE_SORTED_ORACLE_VOTINGS then
                call reset_sorted_oracle_voting_contracts_changes(l_rec.id);
                continue;
            end if;

            if l_rec.type = CHANGE_TYPE_SORTED_ORACLE_VOTING_COMMITTEES then
                call reset_sorted_oracle_voting_contract_committees_changes(l_rec.id);
                continue;
            end if;

            if l_rec.type = CHANGE_TYPE_BALANCE_UPDATE_SUMMARIES then
                call reset_balance_update_summaries_changes(l_rec.id);
                continue;
            end if;

            if l_rec.type = CHANGE_TYPE_MINING_REWARD_SUMMARIES then
                call reset_mining_reward_summaries_changes(l_rec.id);
                continue;
            end if;

        end loop;
    DELETE FROM changes WHERE block_height > p_block_height;
END
$$;