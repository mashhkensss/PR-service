package teamhandler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mashhkensss/PR-service/internal/domain"
	"github.com/mashhkensss/PR-service/internal/domain/requester"
	domainteam "github.com/mashhkensss/PR-service/internal/domain/team"
	domainuser "github.com/mashhkensss/PR-service/internal/domain/user"
	"github.com/mashhkensss/PR-service/internal/http/dto"
	mw "github.com/mashhkensss/PR-service/internal/http/middleware"
)

type teamServiceMock struct {
	addFn func(ctx context.Context, t domainteam.Team) (domainteam.Team, error)
	getFn func(ctx context.Context, name domain.TeamName) (domainteam.Team, error)
	forFn func(ctx context.Context, actor requester.Requester, name domain.TeamName) (domainteam.Team, error)
}

func (m teamServiceMock) AddTeam(ctx context.Context, t domainteam.Team) (domainteam.Team, error) {
	if m.addFn != nil {
		return m.addFn(ctx, t)
	}
	return domainteam.New("backend", nil)
}

func (m teamServiceMock) GetTeam(ctx context.Context, name domain.TeamName) (domainteam.Team, error) {
	if m.getFn != nil {
		return m.getFn(ctx, name)
	}
	return domainteam.New(name, nil)
}

func (m teamServiceMock) GetTeamForUser(ctx context.Context, actor requester.Requester, name domain.TeamName) (domainteam.Team, error) {
	if m.forFn != nil {
		return m.forFn(ctx, actor, name)
	}
	return domainteam.New(name, nil)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAddTeam_Success(t *testing.T) {
	member, _ := domainuser.New("u1", "Alice", "backend", true)
	h := &handler{
		service: teamServiceMock{
			addFn: func(ctx context.Context, aggregate domainteam.Team) (domainteam.Team, error) {
				return domainteam.New(aggregate.TeamName(), aggregate.Members())
			},
		},
		logger: newTestLogger(),
	}

	body := `{"team_name":"backend","members":[{"user_id":"u1","username":"Alice","is_active":true}]}`
	req := httptest.NewRequest(http.MethodPost, "/team/add", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler := mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.AddTeam))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var resp struct {
		Team dto.Team `json:"team"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Team.TeamName != "backend" || len(resp.Team.Members) != 1 || resp.Team.Members[0].UserID != string(member.UserID()) {
		t.Fatalf("unexpected response %+v", resp.Team)
	}
}

func TestAddTeam_InvalidJSON(t *testing.T) {
	h := &handler{service: teamServiceMock{}, logger: newTestLogger()}
	req := httptest.NewRequest(http.MethodPost, "/team/add", strings.NewReader("{"))
	rr := httptest.NewRecorder()
	mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.AddTeam)).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAddTeam_ServiceError(t *testing.T) {
	h := &handler{
		service: teamServiceMock{
			addFn: func(ctx context.Context, t domainteam.Team) (domainteam.Team, error) {
				return domainteam.Team{}, domain.ErrTeamExists
			},
		},
		logger: newTestLogger(),
	}
	body := `{"team_name":"backend","members":[{"user_id":"u1","username":"Alice","is_active":true}]}`
	req := httptest.NewRequest(http.MethodPost, "/team/add", strings.NewReader(body))
	rr := httptest.NewRecorder()
	mw.NewValidatorMiddleware(mw.NewTagValidator())(http.HandlerFunc(h.AddTeam)).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetTeam_Success(t *testing.T) {
	teamAgg, _ := domainteam.New("backend", nil)
	h := &handler{
		service: teamServiceMock{
			forFn: func(ctx context.Context, actor requester.Requester, name domain.TeamName) (domainteam.Team, error) {
				return teamAgg, nil
			},
		},
		logger: newTestLogger(),
	}
	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	rr := httptest.NewRecorder()
	http.HandlerFunc(h.GetTeam).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetTeam_MissingName(t *testing.T) {
	h := &handler{service: teamServiceMock{}, logger: newTestLogger()}
	req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
	rr := httptest.NewRecorder()
	http.HandlerFunc(h.GetTeam).ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetTeam_AccessDenied(t *testing.T) {
	h := &handler{
		service: teamServiceMock{
			forFn: func(ctx context.Context, actor requester.Requester, name domain.TeamName) (domainteam.Team, error) {
				return domainteam.Team{}, domain.ErrTeamAccessDenied
			},
		},
		logger: newTestLogger(),
	}
	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	rr := httptest.NewRecorder()
	http.HandlerFunc(h.GetTeam).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}
