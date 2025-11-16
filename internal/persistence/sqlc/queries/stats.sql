-- CountAssignmentsPerUser
SELECT reviewer_id, COUNT(*) AS assignments
FROM pull_request_reviewers
GROUP BY reviewer_id;

-- CountAssignmentsPerPullRequest
SELECT pull_request_id, COUNT(*) AS assignments
FROM pull_request_reviewers
GROUP BY pull_request_id;
