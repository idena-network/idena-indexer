INSERT INTO BLOCKS (HEIGHT, HASH, EPOCH, TIMESTAMP, IS_EMPTY, VALIDATORS_COUNT, BODY_SIZE, VRF_PROPOSER_THRESHOLD,
                    FULL_SIZE, FEE_RATE)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)