package dify

import (
	"context"
	"errors"
	"net/http"
	"strconv"
)

const (
	FeedbackLike    = "like"
	FeedbackDislike = "dislike"
)

type MessagesFeedbacksRequest struct {
	MessageID string `json:"-"`
	Rating    string `json:"rating"`
	User      string `json:"user"`
	Content   string `json:"content,omitempty"`
}

type MessagesFeedbacksResponse struct {
	HasMore bool                            `json:"has_more"`
	Data    []MessagesFeedbacksDataResponse `json:"data"`
}

type MessagesFeedbacksDataResponse struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	PhoneNumber    string `json:"phone_number"`
	AvatarURL      string `json:"avatar_url"`
	DisplayName    string `json:"display_name"`
	ConversationID string `json:"conversation_id"`
	LastActiveAt   int64  `json:"last_active_at"`
	CreatedAt      int64  `json:"created_at"`
}

type MessagesRequest struct {
	ConversationID string `json:"conversation_id"`
	FirstID        string `json:"first_id,omitempty"`
	Limit          int    `json:"limit"`
	User           string `json:"user"`
}

type MessagesResponse struct {
	Limit   int                    `json:"limit"`
	HasMore bool                   `json:"has_more"`
	Data    []MessagesDataResponse `json:"data"`
}

type MessagesDataResponse struct {
	ID             string         `json:"id"`
	ConversationID string         `json:"conversation_id"`
	Inputs         map[string]any `json:"inputs"`
	Query          string         `json:"query"`
	Answer         string         `json:"answer"`
	Feedback       any            `json:"feedback"`
	CreatedAt      int64          `json:"created_at"`
}

type SuggestedQuestionsResponse struct {
	Result string   `json:"result"`
	Data   []string `json:"data"`
}

type AppFeedbacksRequest struct {
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	User  string `json:"user"`
}

type AppFeedbacksResponse struct {
	HasMore bool   `json:"has_more"`
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Data    []any  `json:"data"`
}

func (api *API) Messages(ctx context.Context, req *MessagesRequest) (resp *MessagesResponse, err error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/messages", nil)
	if err != nil {
		return
	}
	query := httpReq.URL.Query()
	query.Set("conversation_id", req.ConversationID)
	query.Set("user", req.User)
	if req.FirstID != "" {
		query.Set("first_id", req.FirstID)
	}
	if req.Limit > 0 {
		query.Set("limit", strconv.FormatInt(int64(req.Limit), 10))
	}
	httpReq.URL.RawQuery = query.Encode()

	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

func (api *API) MessagesFeedbacks(ctx context.Context, req *MessagesFeedbacksRequest) (resp *MessagesFeedbacksResponse, err error) {
	if req.MessageID == "" {
		err = errors.New("MessagesFeedbacksRequest.MessageID is required")
		return
	}

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/messages/"+req.MessageID+"/feedbacks", req)
	if err != nil {
		return
	}
	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

func (api *API) SuggestedQuestions(ctx context.Context, messageID string, user string) (resp *SuggestedQuestionsResponse, err error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/messages/"+messageID+"/suggested", nil)
	if err != nil {
		return
	}

	query := httpReq.URL.Query()
	query.Set("user", user)
	httpReq.URL.RawQuery = query.Encode()

	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

func (api *API) AppFeedbacks(ctx context.Context, req *AppFeedbacksRequest) (resp *AppFeedbacksResponse, err error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/app/feedbacks", nil)
	if err != nil {
		return
	}

	query := httpReq.URL.Query()
	if req.Page > 0 {
		query.Set("page", strconv.Itoa(req.Page))
	}
	if req.Limit > 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.User != "" {
		query.Set("user", req.User)
	}
	httpReq.URL.RawQuery = query.Encode()

	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}
