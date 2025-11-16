package test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domain "github.com/mashhkensss/PR-service/internal/domain"
	domainpr "github.com/mashhkensss/PR-service/internal/domain/pullrequest"
	domainteam "github.com/mashhkensss/PR-service/internal/domain/team"
	domainuser "github.com/mashhkensss/PR-service/internal/domain/user"
	routerhttp "github.com/mashhkensss/PR-service/internal/http"
	prhandler "github.com/mashhkensss/PR-service/internal/http/handlers/pullrequest"
	statshandler "github.com/mashhkensss/PR-service/internal/http/handlers/stats"
	teamhandler "github.com/mashhkensss/PR-service/internal/http/handlers/team"
	userhandler "github.com/mashhkensss/PR-service/internal/http/handlers/user"
	"github.com/mashhkensss/PR-service/internal/http/middleware"
	"github.com/mashhkensss/PR-service/internal/idempotency"
	"github.com/mashhkensss/PR-service/internal/service/assignment"
	pullrequestservice "github.com/mashhkensss/PR-service/internal/service/pullrequest"
	statsservice "github.com/mashhkensss/PR-service/internal/service/stats"
	teamservice "github.com/mashhkensss/PR-service/internal/service/team"
	userservice "github.com/mashhkensss/PR-service/internal/service/user"
)

func TestEndToEndFlow(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	userRepo := newInMemoryUserRepo()
	teamRepo := newInMemoryTeamRepo(userRepo)
	prRepo := newInMemoryPRRepo()
	statsRepo := newInMemoryStatsRepo(prRepo)

	txRunner := noopTx{}

	teamSvc := teamservice.New(teamRepo, txRunner)
	userSvc := userservice.New(userRepo, prRepo)
	prSvc := pullrequestservice.New(teamRepo, userRepo, prRepo, txRunner, assignment.NewStrategy(nil))
	statsSvc := statsservice.New(statsRepo)

	teamHandler := teamhandler.New(teamSvc, logger.With("handler", "team"))
	userHandler := userhandler.New(userSvc, logger.With("handler", "user"))
	prHandler := prhandler.New(prSvc, logger.With("handler", "pr"))
	statsHandler := statshandler.New(statsSvc, logger.With("handler", "stats"))

	idStore := newMemoryStore()
	auth := middleware.NewAuthorization([]byte("admin-secret"), []byte("user-secret"))
	l := middleware.NewLogger(logger.With("component", "http"))
	idempotencyMW := middleware.NewIdempotencyMiddleware(idStore, time.Minute, logger.With("component", "idempotency"))
	rateLimiter := middleware.NewRateLimiter(100, time.Second, false)
	validator := middleware.NewValidatorMiddleware(middleware.NewTagValidator())

	router := routerhttp.NewRouter(routerhttp.RouterConfig{
		TeamHandler:  teamHandler,
		UserHandler:  userHandler,
		PRHandler:    prHandler,
		StatsHandler: statsHandler,
		Auth:         auth,
		Logger:       l.Middleware,
		RateLimiter:  rateLimiter.Middleware,
		Idempotency:  idempotencyMW,
		Validator:    validator,
	})

	adminToken := makeToken("admin-secret", "admin", "admin")

	teamBody := `{"team_name":"backend","members":[{"user_id":"author","username":"Alice","is_active":true},{"user_id":"rev1","username":"Bob","is_active":true},{"user_id":"rev2","username":"Charlie","is_active":true}]}`
	doRequest(t, router, stdhttp.MethodPost, "/team/add", adminToken, teamBody, stdhttp.StatusCreated)

	createBody := `{"pull_request_id":"pr-1","pull_request_name":"Feature","author_id":"author"}`
	doRequest(t, router, stdhttp.MethodPost, "/pullRequest/create", adminToken, createBody, stdhttp.StatusCreated)

	mergeBody := `{"pull_request_id":"pr-1"}`
	doRequest(t, router, stdhttp.MethodPost, "/pullRequest/merge", adminToken, mergeBody, stdhttp.StatusOK)

	userToken := makeToken("user-secret", "rev1", "user")
	resp := doRequest(t, router, stdhttp.MethodGet, "/users/getReview?user_id=rev1", userToken, "", stdhttp.StatusOK)
	var assignments struct {
		PullRequests []struct {
			PullRequestID string `json:"pull_request_id"`
		} `json:"pull_requests"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&assignments); err != nil {
		t.Fatalf("decode assignments: %v", err)
	}
	if len(assignments.PullRequests) != 1 {
		t.Fatalf("expected 1 assignment, got %v", assignments.PullRequests)
	}

	doRequest(t, router, stdhttp.MethodGet, "/stats/summary", adminToken, "", stdhttp.StatusOK)
}

func doRequest(t *testing.T, handler stdhttp.Handler, method, path, token, body string, expected int) *stdhttp.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if method == stdhttp.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	resp := rr.Result()
	if resp.StatusCode != expected {
		t.Fatalf("expected status %d for %s %s, got %d", expected, method, path, resp.StatusCode)
	}
	return resp
}

func makeToken(secret, sub, role string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"` + sub + `","role":"` + role + `","exp":` + fmt.Sprint(time.Now().Add(time.Hour).Unix()) + `}`))
	data := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return data + "." + signature
}

type noopTx struct{}

func (noopTx) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type inMemoryUserRepo struct {
	users map[domain.UserID]domainuser.User
}

func newInMemoryUserRepo() *inMemoryUserRepo {
	return &inMemoryUserRepo{users: make(map[domain.UserID]domainuser.User)}
}

func (r *inMemoryUserRepo) upsert(user domainuser.User) {
	r.users[user.UserID()] = user
}

func (r *inMemoryUserRepo) SetUserActivity(ctx context.Context, id domain.UserID, active bool) (domainuser.User, error) {
	u, ok := r.users[id]
	if !ok {
		return domainuser.User{}, fmt.Errorf("user not found")
	}
	r.users[id] = u.WithActivity(active)
	return r.users[id], nil
}

func (r *inMemoryUserRepo) GetUser(ctx context.Context, id domain.UserID) (domainuser.User, error) {
	u, ok := r.users[id]
	if !ok {
		return domainuser.User{}, fmt.Errorf("user not found")
	}
	return u, nil
}

type inMemoryTeamRepo struct {
	teams map[domain.TeamName]domainteam.Team
	users *inMemoryUserRepo
}

func newInMemoryTeamRepo(users *inMemoryUserRepo) *inMemoryTeamRepo {
	return &inMemoryTeamRepo{
		teams: make(map[domain.TeamName]domainteam.Team),
		users: users,
	}
}

func (r *inMemoryTeamRepo) SaveTeam(ctx context.Context, t domainteam.Team) error {
	if _, exists := r.teams[t.TeamName()]; exists {
		return domain.ErrTeamExists
	}
	r.teams[t.TeamName()] = t
	for _, member := range t.Members() {
		r.users.upsert(member)
	}
	return nil
}

func (r *inMemoryTeamRepo) GetTeam(ctx context.Context, name domain.TeamName) (domainteam.Team, error) {
	t, ok := r.teams[name]
	if !ok {
		return domainteam.Team{}, fmt.Errorf("team not found")
	}
	return t, nil
}

type inMemoryPRRepo struct {
	prs map[domain.PullRequestID]domainpr.PullRequest
}

func newInMemoryPRRepo() *inMemoryPRRepo {
	return &inMemoryPRRepo{prs: make(map[domain.PullRequestID]domainpr.PullRequest)}
}

func (r *inMemoryPRRepo) CreatePullRequest(ctx context.Context, pr domainpr.PullRequest) error {
	if _, exists := r.prs[pr.PullRequestID()]; exists {
		return domain.ErrPullRequestExists
	}
	r.prs[pr.PullRequestID()] = pr
	return nil
}

func (r *inMemoryPRRepo) UpdatePullRequest(ctx context.Context, pr domainpr.PullRequest) error {
	r.prs[pr.PullRequestID()] = pr
	return nil
}

func (r *inMemoryPRRepo) GetPullRequest(ctx context.Context, id domain.PullRequestID) (domainpr.PullRequest, error) {
	pr, ok := r.prs[id]
	if !ok {
		return domainpr.PullRequest{}, fmt.Errorf("not found")
	}
	return pr, nil
}

func (r *inMemoryPRRepo) GetPullRequestForUpdate(ctx context.Context, id domain.PullRequestID) (domainpr.PullRequest, error) {
	return r.GetPullRequest(ctx, id)
}

func (r *inMemoryPRRepo) ListPullRequestsByReviewer(ctx context.Context, reviewer domain.UserID) ([]domainpr.PullRequest, error) {
	result := make([]domainpr.PullRequest, 0)
	for _, pr := range r.prs {
		for _, assigned := range pr.AssignedReviewers() {
			if assigned == reviewer {
				result = append(result, pr)
				break
			}
		}
	}
	return result, nil
}

type inMemoryStatsRepo struct {
	prs *inMemoryPRRepo
}

func newInMemoryStatsRepo(prs *inMemoryPRRepo) *inMemoryStatsRepo {
	return &inMemoryStatsRepo{prs: prs}
}

func (r *inMemoryStatsRepo) AssignmentsPerUser(ctx context.Context) (map[domain.UserID]int, error) {
	result := make(map[domain.UserID]int)
	for _, pr := range r.prs.prs {
		for _, reviewer := range pr.AssignedReviewers() {
			result[reviewer]++
		}
	}
	return result, nil
}

func (r *inMemoryStatsRepo) AssignmentsPerPullRequest(ctx context.Context) (map[domain.PullRequestID]int, error) {
	result := make(map[domain.PullRequestID]int)
	for id, pr := range r.prs.prs {
		result[id] = len(pr.AssignedReviewers())
	}
	return result, nil
}

type memoryStore struct {
	data map[string]idempotency.StoredResponse
	reqs map[string]idempotency.StoredRequest
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		data: make(map[string]idempotency.StoredResponse),
		reqs: make(map[string]idempotency.StoredRequest),
	}
}

func (m *memoryStore) Get(ctx context.Context, key string) (idempotency.StoredRequest, idempotency.StoredResponse, bool, error) {
	req, ok := m.reqs[key]
	if !ok {
		return idempotency.StoredRequest{}, idempotency.StoredResponse{}, false, nil
	}
	return req, m.data[key], true, nil
}

func (m *memoryStore) Save(ctx context.Context, key string, req idempotency.StoredRequest, ttl time.Duration, resp idempotency.StoredResponse) error {
	m.reqs[key] = req
	m.data[key] = resp
	return nil
}
