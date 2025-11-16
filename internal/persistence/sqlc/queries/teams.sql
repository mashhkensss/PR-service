-- UpsertTeam
INSERT INTO teams (team_name)
VALUES ($1)
ON CONFLICT (team_name) DO UPDATE SET updated_at = NOW();

-- UpsertUser
INSERT INTO users (user_id, username, team_name, is_active)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) DO UPDATE
SET username = EXCLUDED.username,
    team_name = EXCLUDED.team_name,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- GetTeamWithMembers
SELECT
    t.team_name,
    u.user_id,
    u.username,
    u.is_active
FROM teams t
LEFT JOIN users u ON u.team_name = t.team_name
WHERE t.team_name = $1
ORDER BY u.username, u.user_id;
