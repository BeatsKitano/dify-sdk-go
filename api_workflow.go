package dify

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	EventWorkflowStarted  = "workflow_started"
	EventNodeStarted      = "node_started"
	EventNodeFinished     = "node_finished"
	EventWorkflowFinished = "workflow_finished"
	EventTextChunk        = "text_chunk"
	EventTTSMessage       = "tts_message"
	EventTTSMessageEnd    = "tts_message_end"
)

type FileInput struct {
	Type           string `json:"type"`
	TransferMethod string `json:"transfer_method"`
	URL            string `json:"url,omitempty"`
	UploadFileID   string `json:"upload_file_id,omitempty"`
}

type WorkflowRequest struct {
	Inputs       map[string]any `json:"inputs"`
	ResponseMode string         `json:"response_mode"`
	User         string         `json:"user"`
	Files        []FileInput    `json:"files,omitempty"`
}

type WorkflowResponse struct {
	WorkflowRunID string `json:"workflow_run_id"`
	TaskID        string `json:"task_id"`
	Data          struct {
		ID          string         `json:"id"`
		WorkflowID  string         `json:"workflow_id"`
		Status      string         `json:"status"`
		Outputs     map[string]any `json:"outputs"`
		Error       string         `json:"error,omitempty"`
		ElapsedTime float64        `json:"elapsed_time"`
		TotalTokens int            `json:"total_tokens"`
		TotalSteps  int            `json:"total_steps"`
		CreatedAt   int64          `json:"created_at"`
		FinishedAt  int64          `json:"finished_at"`
	} `json:"data"`
}

type StreamingResponse struct {
	Event          string `json:"event"`
	TaskID         string `json:"task_id"`
	WorkflowRunID  string `json:"workflow_run_id"`
	SequenceNumber int    `json:"sequence_number"`
	Data           struct {
		ID                string         `json:"id"`
		WorkflowID        string         `json:"workflow_id,omitempty"`
		NodeID            string         `json:"node_id,omitempty"`
		NodeType          string         `json:"node_type,omitempty"`
		Title             string         `json:"title,omitempty"`
		Index             int            `json:"index"`
		Predecessor       string         `json:"predecessor_node_id,omitempty"`
		Inputs            []any          `json:"inputs,omitempty"`
		Outputs           map[string]any `json:"outputs,omitempty"`
		Status            string         `json:"status,omitempty"`
		Error             string         `json:"error,omitempty"`
		ElapsedTime       float64        `json:"elapsed_time,omitempty"`
		ExecutionMetadata struct {
			TotalTokens int     `json:"total_tokens,omitempty"`
			TotalPrice  float64 `json:"total_price,omitempty"`
			Currency    string  `json:"currency,omitempty"`
		} `json:"execution_metadata,omitempty"`
		CreatedAt  int64 `json:"created_at"`
		FinishedAt int64 `json:"finished_at,omitempty"`
	} `json:"data"`
}

type TTSMessage struct {
	Event     string `json:"event"`
	TaskID    string `json:"task_id"`
	MessageID string `json:"message_id"`
	Audio     string `json:"audio"`
	CreatedAt int64  `json:"created_at"`
}

type TextChunkResponse struct {
	Event     string `json:"event"`
	TaskID    string `json:"task_id"`
	MessageID string `json:"message_id"`
	Text      string `json:"text"`
	CreatedAt int64  `json:"created_at"`
}

type EventHandler interface {
	HandleStreamingResponse(StreamingResponse)
	HandleTTSMessage(TTSMessage)
	HandleTextChunk(TextChunkResponse)
}

type DefaultEventHandler struct {
	StreamHandler   func(StreamingResponse)
	TTSHandler      func(TTSMessage)
	TextChunkHandler func(TextChunkResponse)
}

func (h *DefaultEventHandler) HandleStreamingResponse(resp StreamingResponse) {
	if h.StreamHandler != nil {
		h.StreamHandler(resp)
	}
}

func (h *DefaultEventHandler) HandleTTSMessage(msg TTSMessage) {
	if h.TTSHandler != nil {
		h.TTSHandler(msg)
	}
}

func (h *DefaultEventHandler) HandleTextChunk(msg TextChunkResponse) {
	if h.TextChunkHandler != nil {
		h.TextChunkHandler(msg)
	}
}

type WorkflowLogsRequest struct {
	Keyword string `json:"keyword,omitempty"`
	Status  string `json:"status,omitempty"`
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
}

type WorkflowLogsResponse struct {
	Page    int                  `json:"page"`
	Limit   int                  `json:"limit"`
	Total   int                  `json:"total"`
	HasMore bool                 `json:"has_more"`
	Data    []WorkflowLogsData   `json:"data"`
}

type WorkflowLogsData struct {
	ID                string  `json:"id"`
	WorkflowRunID     string  `json:"workflow_run_id"`
	Status            string  `json:"status"`
	ElapsedTime       float64 `json:"elapsed_time"`
	TotalTokens       int     `json:"total_tokens"`
	TotalSteps        int     `json:"total_steps"`
	CreatedAt         int64   `json:"created_at"`
	FinishedAt        int64   `json:"finished_at"`
	CreatedByAccount  string  `json:"created_by_account"`
}

func (api *API) RunWorkflow(ctx context.Context, request WorkflowRequest) (*WorkflowResponse, error) {
	request.ResponseMode = "blocking"

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/workflows/run", request)
	if err != nil {
		return nil, err
	}

	var workflowResp WorkflowResponse
	if err := api.c.sendJSONRequest(httpReq, &workflowResp); err != nil {
		return nil, err
	}
	return &workflowResp, nil
}

func (api *API) RunStreamWorkflow(ctx context.Context, request WorkflowRequest, handler func(StreamingResponse)) error {
	return api.RunStreamWorkflowWithHandler(ctx, request, &DefaultEventHandler{StreamHandler: handler})
}

func (api *API) RunStreamWorkflowWithHandler(ctx context.Context, request WorkflowRequest, handler EventHandler) error {
	request.ResponseMode = "streaming"

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/workflows/run", request)
	if err != nil {
		return err
	}

	resp, err := api.c.sendRequest(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %s: %s", resp.Status, readResponseBody(resp.Body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading streaming response: %w", err)
		}

		if len(line) > 6 && string(line[:6]) == "data: " {
			var event struct {
				Event string `json:"event"`
			}
			if err := json.Unmarshal(line[6:], &event); err != nil {
				continue
			}

			switch event.Event {
			case EventTTSMessage, EventTTSMessageEnd:
				var ttsMsg TTSMessage
				if err := json.Unmarshal(line[6:], &ttsMsg); err != nil {
					continue
				}
				handler.HandleTTSMessage(ttsMsg)

			case EventTextChunk:
				var textChunk TextChunkResponse
				if err := json.Unmarshal(line[6:], &textChunk); err != nil {
					continue
				}
				handler.HandleTextChunk(textChunk)

			default:
				var streamResp StreamingResponse
				if err := json.Unmarshal(line[6:], &streamResp); err != nil {
					continue
				}
				handler.HandleStreamingResponse(streamResp)
			}
		}
	}

	return nil
}

func (api *API) GetWorkflowRun(ctx context.Context, workflowRunID string) (*WorkflowResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/workflows/run/"+workflowRunID, nil)
	if err != nil {
		return nil, err
	}

	var resp WorkflowResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) WorkflowLogs(ctx context.Context, req *WorkflowLogsRequest) (*WorkflowLogsResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/workflows/logs", nil)
	if err != nil {
		return nil, err
	}

	query := httpReq.URL.Query()
	if req.Keyword != "" {
		query.Set("keyword", req.Keyword)
	}
	if req.Status != "" {
		query.Set("status", req.Status)
	}
	if req.Page > 0 {
		query.Set("page", strconv.Itoa(req.Page))
	}
	if req.Limit > 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	httpReq.URL.RawQuery = query.Encode()

	var resp WorkflowLogsResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) StopWorkflowTask(ctx context.Context, taskID string, user string) error {
	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/workflows/tasks/"+taskID+"/stop", map[string]string{
		"user": user,
	})
	if err != nil {
		return err
	}
	return api.c.sendJSONRequest(httpReq, nil)
}

func readResponseBody(body io.Reader) string {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return fmt.Sprintf("failed to read response body: %v", err)
	}
	return string(bodyBytes)
}
