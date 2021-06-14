package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	diff                  = "diff"
	defaultLocalSecureURI = "https://localhost:8443"
	sync                  = "sync"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client represents an API client.
type Client struct {
	authToken  string
	httpClient httpClient
	endpoint   string
}

// NewClient returns a new API client.
func NewClient(endpoint, authToken string) Client {
	// Automatically disable TLS verification if it's a local endpoint.
	// TODO handle this better.
	tr := &http.Transport{}
	if endpoint == defaultLocalSecureURI {
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return Client{
		authToken:  authToken,
		endpoint:   endpoint,
		httpClient: &http.Client{Transport: tr},
	}
}

// TODO rename
// GetWorkflowStatusResponse represents the status of a workflow.
type GetWorkflowStatusResponse struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Finished string `json:"finished"`
}

// TODO rename
// GetWorkflowsResponse represents a collection of workflows.
type GetWorkflowsResponse []string

// TODO rename
// targetOperation represents the request for either a 'diff' or 'sync' against
// a project target.
type targetOperationRequest struct {
	Path string `json:"path"`
	SHA  string `json:"sha"`
	Type string `json:"type"`
}

// TODO rename
// targetOperationResponse represents a response to a target operation.
type targetOperationResponse struct {
	WorkflowName string `json:"workflow_name"`
}

// TODO rename
// DiffResponse represents a diff response.
type DiffResponse targetOperationResponse

// ExecuteWorkflowInput represents the input for ExecuteWorkflow.
type ExecuteWorkflowInput struct {
	Arguments            map[string][]string `json:"arguments"`
	EnvironmentVariables map[string]string   `json:"environment_variables"`
	Framework            string              `json:"framework"`
	Parameters           map[string]string   `json:"parameters"`
	ProjectName          string              `json:"project_name"`
	TargetName           string              `json:"target_name"`
	Type                 string              `json:"type"`
	WorkflowTemplateName string              `json:"workflow_template_name"`
}

// ExecuteWorkflowOutput represents the output for ExecuteWorkflow.
type ExecuteWorkflowOutput struct {
	WorkflowName string `json:"workflow_name"`
}

// TODO rename
// SyncResponse represents a sync response.
type SyncResponse targetOperationResponse

// GetWorkflowStatus gets the status of a workflow.
func (c *Client) GetWorkflowStatus(ctx context.Context, name string) (GetWorkflowStatusResponse, error) {
	url := fmt.Sprintf("%s/workflows/%s", c.endpoint, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return GetWorkflowStatusResponse{}, fmt.Errorf("unable to create api request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GetWorkflowStatusResponse{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GetWorkflowStatusResponse{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		return GetWorkflowStatusResponse{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// TODO rename
	var wfResp GetWorkflowStatusResponse
	if err := json.Unmarshal(body, &wfResp); err != nil {
		return GetWorkflowStatusResponse{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return wfResp, nil
}

// GetWorkflows gets the list of workflows for a project and target.
func (c *Client) GetWorkflows(ctx context.Context, project, target string) (GetWorkflowsResponse, error) {
	url := fmt.Sprintf("%s/projects/%s/targets/%s/workflows", c.endpoint, project, target)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return GetWorkflowsResponse{}, fmt.Errorf("unable to create api request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GetWorkflowsResponse{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GetWorkflowsResponse{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		return GetWorkflowsResponse{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// TODO rename
	var wfResp GetWorkflowsResponse
	if err := json.Unmarshal(body, &wfResp); err != nil {
		return GetWorkflowsResponse{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return wfResp, nil
}

// Diff submits a "diff" for the provided project target.
func (c *Client) Diff(ctx context.Context, project, target, sha, path string) (DiffResponse, error) {
	url := fmt.Sprintf("%s/projects/%s/targets/%s/operations", c.endpoint, project, target)

	targetReq := targetOperationRequest{
		Path: path,
		SHA:  sha,
		Type: diff,
	}

	reqBody, err := json.Marshal(targetReq)
	if err != nil {
		return DiffResponse{}, fmt.Errorf("unable to create api request body, error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return DiffResponse{}, fmt.Errorf("unable to create api request: %w", err)
	}

	req.Header.Add("Authorization", c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return DiffResponse{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return DiffResponse{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return DiffResponse{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// TOOD rename
	var diffResp DiffResponse
	if err := json.Unmarshal(body, &diffResp); err != nil {
		return DiffResponse{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return diffResp, nil
}

// ExecuteWorkflow submits a workflow execution request.
func (c *Client) ExecuteWorkflow(ctx context.Context, input ExecuteWorkflowInput) (ExecuteWorkflowOutput, error) {
	// TODO this should probably be refactored to be a different operation type
	// (like diff/sync).
	url := fmt.Sprintf("%s/workflows", c.endpoint)

	reqBody, err := json.Marshal(input)
	if err != nil {
		return ExecuteWorkflowOutput{}, fmt.Errorf("unable to create api request body, error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return ExecuteWorkflowOutput{}, fmt.Errorf("unable to create api request: %w", err)
	}

	req.Header.Add("Authorization", c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ExecuteWorkflowOutput{}, fmt.Errorf("unable to make api call: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ExecuteWorkflowOutput{}, fmt.Errorf("error reading response body. status code: %d, error: %w", resp.StatusCode, err)
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return ExecuteWorkflowOutput{}, fmt.Errorf("received unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var output ExecuteWorkflowOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return ExecuteWorkflowOutput{}, fmt.Errorf("unable to parse response: %w", err)
	}

	return output, nil
}
