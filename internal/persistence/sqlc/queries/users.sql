-- UpdateUserActivity
UPDATE users
SET is_active = $2,
    updated_at = NOW()
WHERE user_id = $1
RETURNING user_id, username, team_name, is_active;

-- GetUser
SELECT user_id, username, team_name, is_active
FROM users
WHERE user_id = $1;
