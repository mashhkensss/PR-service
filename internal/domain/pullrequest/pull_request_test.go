package pullrequest

import (
	"testing"
	"time"

	"github.com/mashhkensss/PR-service/internal/domain"
)

func TestNewPullRequestDefaults(t *testing.T) {
	now := time.Now().Add(-time.Hour)
	pr, err := New("pr-1", " Feature ", "u1", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr.PullRequestID() != "pr-1" {
		t.Fatalf("id mismatch: %s", pr.PullRequestID())
	}
	if pr.PullRequestName() != "Feature" {
		t.Fatalf("name must be trimmed")
	}
	if pr.Status() != domain.PullRequestStatusOpen {
		t.Fatalf("default status must be OPEN")
	}
	if pr.CreatedAt().IsZero() {
		t.Fatalf("createdAt should be set")
	}
	if len(pr.AssignedReviewers()) != 0 {
		t.Fatalf("expected empty reviewers")
	}
}

func TestAssignReviewersNormalizesList(t *testing.T) {
	pr, _ := New("pr-1", "Feature", "author", time.Time{})
	err := pr.AssignReviewers([]domain.UserID{"rev1", "", "rev1", "rev2", "rev3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	reviewers := pr.AssignedReviewers()
	if len(reviewers) != 2 || reviewers[0] != "rev1" || reviewers[1] != "rev2" {
		t.Fatalf("unexpected reviewers list: %v", reviewers)
	}
}

func TestAppendReviewerConstraints(t *testing.T) {
	pr, _ := New("pr-1", "Feature", "author", time.Time{})
	if err := pr.AppendReviewer("author"); err != domain.ErrAuthorIsReviewer {
		t.Fatalf("expected ErrAuthorIsReviewer, got %v", err)
	}
	if err := pr.AppendReviewer("rev1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := pr.AppendReviewer("rev1"); err != domain.ErrReviewerAlreadyAssigned {
		t.Fatalf("expected ErrReviewerAlreadyAssigned, got %v", err)
	}
}

func TestReplaceReviewer(t *testing.T) {
	pr, _ := New("pr-1", "Feature", "author", time.Time{})
	_ = pr.AssignReviewers([]domain.UserID{"rev1", "rev2"})
	if err := pr.ReplaceReviewer("rev1", "rev2"); err != domain.ErrReviewerAlreadyAssigned {
		t.Fatalf("expected ErrReviewerAlreadyAssigned, got %v", err)
	}
	if err := pr.ReplaceReviewer("unknown", "rev3"); err != domain.ErrReviewerNotAssigned {
		t.Fatalf("expected ErrReviewerNotAssigned, got %v", err)
	}
	if err := pr.ReplaceReviewer("rev1", "author"); err != domain.ErrAuthorIsReviewer {
		t.Fatalf("expected ErrAuthorIsReviewer, got %v", err)
	}
	if err := pr.ReplaceReviewer("rev1", "rev3"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	final := pr.AssignedReviewers()
	if final[0] != "rev3" {
		t.Fatalf("expected rev3 in slot 0, got %v", final)
	}
}

func TestMergeIdempotent(t *testing.T) {
	pr, _ := New("pr-1", "Feature", "author", time.Time{})
	if changed := pr.Merge(time.Now()); !changed {
		t.Fatalf("first merge should change state")
	}
	if changed := pr.Merge(time.Now()); changed {
		t.Fatalf("second merge must be idempotent")
	}
}
