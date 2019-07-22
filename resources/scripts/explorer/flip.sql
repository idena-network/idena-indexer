select f.id,
       coalesce(f.answer, '') answer,
       coalesce(f.status, '') status,
       f.data
from flips f
where LOWER(f.cid) = LOWER($1)