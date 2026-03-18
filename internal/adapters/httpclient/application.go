package httpclient

import (
	"context"
	"fmt"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// ApplicationClient implements ports.ApplicationService against the
// /api/v1/jobs and /api/v1/pipeline endpoints.
//
// All methods require a valid Bearer token in the underlying Client.
// The userID parameter is accepted for interface compatibility but is ignored —
// the server resolves the user from the Authorization header.
type ApplicationClient struct {
	c *Client
}

// NewApplicationClient constructs an ApplicationClient.
func NewApplicationClient(c *Client) *ApplicationClient {
	return &ApplicationClient{c: c}
}

// SetStatus calls the appropriate POST /api/v1/jobs/{id}/<action> endpoint
// based on the target status.
func (a *ApplicationClient) SetStatus(ctx context.Context, _ domain.UserID, jobID domain.JobID, status domain.ApplicationStatus) error {
	id := jobID.String()

	switch status {
	case domain.StatusInterested:
		if err := a.c.post(ctx, "jobs/"+id+"/interested", true, nil, nil); err != nil {
			return fmt.Errorf("mark interested: %w", err)
		}
	case domain.StatusApplied:
		if err := a.c.post(ctx, "jobs/"+id+"/apply", true, nil, nil); err != nil {
			return fmt.Errorf("mark applied: %w", err)
		}
	default:
		body := map[string]string{"status": string(status)}
		if err := a.c.post(ctx, "jobs/"+id+"/status", true, body, nil); err != nil {
			return fmt.Errorf("set status: %w", err)
		}
	}

	return nil
}

// SetNotes calls POST /api/v1/jobs/{id}/notes.
func (a *ApplicationClient) SetNotes(ctx context.Context, _ domain.UserID, jobID domain.JobID, notes string) error {
	body := map[string]string{"notes": notes}
	if err := a.c.post(ctx, "jobs/"+jobID.String()+"/notes", true, body, nil); err != nil {
		return fmt.Errorf("set notes: %w", err)
	}
	return nil
}

// GetUserJob is not used by the CLI and is not exposed by the JSON API.
func (a *ApplicationClient) GetUserJob(_ context.Context, _ domain.UserID, _ domain.JobID) (domain.UserJob, error) {
	return domain.UserJob{}, fmt.Errorf("GetUserJob: not supported in remote mode")
}

// ListByStatus calls GET /api/v1/jobs/interested or /api/v1/jobs/applied.
func (a *ApplicationClient) ListByStatus(ctx context.Context, _ domain.UserID, status domain.ApplicationStatus) ([]domain.Job, error) {
	var path string
	switch status {
	case domain.StatusInterested:
		path = "jobs/interested"
	case domain.StatusApplied:
		path = "jobs/applied"
	default:
		return nil, fmt.Errorf("ListByStatus: status %q not supported in remote mode", status)
	}

	var jobs []domain.Job
	if err := a.c.get(ctx, path, nil, true, &jobs); err != nil {
		return nil, fmt.Errorf("list by status: %w", err)
	}
	if jobs == nil {
		jobs = []domain.Job{}
	}
	return jobs, nil
}

// pipelineGroup mirrors the JSON shape returned by GET /api/v1/pipeline.
type pipelineGroup struct {
	Status string       `json:"status"`
	Jobs   []domain.Job `json:"jobs"`
}

// ListPipeline calls GET /api/v1/pipeline and converts the response into the
// map[ApplicationStatus][]Job shape expected by the port.
func (a *ApplicationClient) ListPipeline(ctx context.Context, _ domain.UserID) (map[domain.ApplicationStatus][]domain.Job, error) {
	var groups []pipelineGroup
	if err := a.c.get(ctx, "pipeline", nil, true, &groups); err != nil {
		return nil, fmt.Errorf("list pipeline: %w", err)
	}

	out := make(map[domain.ApplicationStatus][]domain.Job, len(groups))
	for _, g := range groups {
		out[domain.ApplicationStatus(g.Status)] = g.Jobs
	}
	return out, nil
}
