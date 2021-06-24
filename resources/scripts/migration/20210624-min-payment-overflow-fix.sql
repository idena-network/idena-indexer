ALTER TABLE oracle_voting_contracts
    ALTER COLUMN voting_min_payment TYPE numeric(48, 18);
ALTER TABLE oracle_voting_contract_call_starts
    ALTER COLUMN voting_min_payment TYPE numeric(48, 18);
DROP FUNCTION calculate_estimated_oracle_reward;