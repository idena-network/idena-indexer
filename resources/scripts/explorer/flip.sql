select f.id,
       coalesce(f.answer, '') answer,
       coalesce(f.status, '') status
from flips f
where LOWER(f.cid) = LOWER($1)