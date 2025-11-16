-- InsertPullRequest
INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- ReplacePullRequestReviewers
DELETE FROM pull_request_reviewers
WHERE pull_request_id = $1;

-- InsertPullRequestReviewer
INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id, slot)
VALUES ($1, $2, $3);

-- GetPullRequest
SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at, updated_at
FROM pull_requests
WHERE pull_request_id = $1;

-- ListReviewers
SELECT reviewer_id
FROM pull_request_reviewers
WHERE pull_request_id = $1
ORDER BY slot ASC;

-- ListPullRequestsByReviewer
SELECT pr.pull_request_id,
       pr.pull_request_name,
       pr.author_id,
       pr.status,
       pr.created_at,
       pr.merged_at,
       pr.updated_at
FROM pull_requests pr
JOIN pull_request_reviewers r ON r.pull_request_id = pr.pull_request_id
WHERE r.reviewer_id = $1
ORDER BY pr.created_at DESC;
