package dify

import (
	"context"
	"errors"
	"net/http"
	"strconv"
)

type ConversationsRequest struct {
	LastID string `json:"last_id,omitempty"`
	Limit  int    `json:"limit"`
	SortBy string `json:"sort_by,omitempty"`
	User   string `json:"user"`
}

type ConversationsResponse struct {
	Limit   int                         `json:"limit"`
	HasMore bool                        `json:"has_more"`
	Data    []ConversationsDataResponse `json:"data"`
}

type ConversationsDataResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Inputs    map[string]any    `json:"inputs"`
	Status    string            `json:"status"`
	CreatedAt int64             `json:"created_at"`
}

type ConversationsRenamingRequest struct {
	ConversationID string `json:"-"`
	Name           string `json:"name"`
	AutoGenerate   *bool  `json:"auto_generate,omitempty"`
	User           string `json:"user"`
}

type ConversationsRenamingResponse struct {
	Result string `json:"result"`
}

func (api *API) Conversations(ctx context.Context, req *ConversationsRequest) (resp *ConversationsResponse, err error) {
	if req.User == "" {
		err = errors.New("ConversationsRequest.User is required")
		return
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/conversations", nil)
	if err != nil {
		return
	}

	query := httpReq.URL.Query()
	query.Set("user", req.User)
	query.Set("limit", strconv.FormatInt(int64(req.Limit), 10))
	if req.LastID != "" {
		query.Set("last_id", req.LastID)
	}
	if req.SortBy != "" {
		query.Set("sort_by", req.SortBy)
	}
	httpReq.URL.RawQuery = query.Encode()

	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

func (api *API) ConversationsRenaming(ctx context.Context, req *ConversationsRenamingRequest) (resp *ConversationsRenamingResponse, err error) {
	url := "/v1/conversations/" + req.ConversationID + "/name"

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return
	}
	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

func (api *API) DeleteConversation(ctx context.Context, conversationID string, user string) error {
	httpReq, err := api.createBaseRequest(ctx, http.MethodDelete, "/v1/conversations/"+conversationID, nil)
	if err != nil {
		return err
	}

	query := httpReq.URL.Query()
	query.Set("user", user)
	httpReq.URL.RawQuery = query.Encode()

	return api.c.sendDeleteRequest(httpReq)
}
