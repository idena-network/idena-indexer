select address, balance, stake from current_balances
order by balance desc
limit $2
    offset $1