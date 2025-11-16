package teamrepo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/mashhkensss/PR-service/internal/domain"
	domainteam "github.com/mashhkensss/PR-service/internal/domain/team"
	domainuser "github.com/mashhkensss/PR-service/internal/domain/user"
	"github.com/mashhkensss/PR-service/internal/persistence/postgres"
)

type Repository struct {
	db  *sql.DB
	sql sq.StatementBuilderType
}

func New(db *sql.DB) *Repository {
	return &Repository{
		db:  db,
		sql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *Repository) SaveTeam(ctx context.Context, aggregate domainteam.Team) error {
	exec := postgres.ExecutorFromContext(ctx, r.db)
	query, args, err := r.sql.Insert("teams").
		Columns("team_name", "updated_at").
		Values(aggregate.TeamName(), time.Now().UTC()).
		Suffix("ON CONFLICT DO NOTHING").
		ToSql()
	if err != nil {
		return err
	}

	res, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("upsert team: %w", err)
	}
	if rows, err := res.RowsAffected(); err == nil && rows == 0 {
		return domain.ErrTeamExists
	}

	for _, member := range aggregate.Members() {
		userQuery, userArgs, err := r.sql.Insert("users").
			Columns("user_id", "username", "team_name", "is_active", "updated_at").
			Values(member.UserID(), member.Username(), aggregate.TeamName(), member.IsActive(), time.Now().UTC()).
			Suffix("ON CONFLICT (user_id) DO UPDATE SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active, updated_at = EXCLUDED.updated_at").
			ToSql()

		if err != nil {
			return err
		}

		if _, err := exec.ExecContext(ctx, userQuery, userArgs...); err != nil {
			return fmt.Errorf("upsert user %s: %w", member.UserID(), err)
		}
	}
	return nil
}

func (r *Repository) GetTeam(ctx context.Context, name domain.TeamName) (domainteam.Team, error) {
	exec := postgres.ExecutorFromContext(ctx, r.db)

	query, args, err := r.sql.Select("t.team_name", "u.user_id", "u.username", "u.is_active").
		From("teams t").
		LeftJoin("users u ON u.team_name = t.team_name").
		Where("t.team_name = ?", name).
		OrderBy("u.username", "u.user_id").
		ToSql()
	if err != nil {
		return domainteam.Team{}, err
	}

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return domainteam.Team{}, fmt.Errorf("query team: %w", err)
	}

	defer rows.Close()

	members := make([]domainuser.User, 0)
	found := false
	var teamName string

	for rows.Next() {
		var (
			teamVal string
			userID  sql.NullString
			nameVal sql.NullString
			active  sql.NullBool
		)
		if err := rows.Scan(&teamVal, &userID, &nameVal, &active); err != nil {
			return domainteam.Team{}, fmt.Errorf("scan team row: %w", err)
		}
		found = true
		teamName = teamVal
		if !userID.Valid {
			continue
		}
		member, err := domainuser.New(domain.UserID(userID.String), nameVal.String, domain.TeamName(teamVal), active.Bool)
		if err != nil {
			return domainteam.Team{}, fmt.Errorf("build user: %w", err)
		}

		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return domainteam.Team{}, fmt.Errorf("rows error: %w", err)
	}

	if !found {
		return domainteam.Team{}, sql.ErrNoRows
	}

	return domainteam.New(domain.TeamName(teamName), members)
}
