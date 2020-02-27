UPDATE FLIPS
SET SIZE=$1
WHERE lower(CID) = lower($2)
  and SIZE = 0