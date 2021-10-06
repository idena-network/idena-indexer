ALTER TABLE epoch_identities
    ADD COLUMN address_id bigint;

UPDATE epoch_identities
SET address_id=s.address_id
FROM (SELECT id, address_id
      FROM address_states) s
WHERE epoch_identities.address_state_id = s.id;

ALTER TABLE epoch_identities
    ALTER COLUMN address_id SET NOT NULL;