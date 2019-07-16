select f.id,
       coalesce(f.answer, '') answer,
       coalesce(f.status, '') status
from flips f
where f.cid = $1