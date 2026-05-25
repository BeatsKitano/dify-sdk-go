package dify

import (
	"context"
	"net/http"
)

type ChatMessageRequest struct {
	Inputs           map[string]interface{} `json:"inputs"`
	Query            string                 `json:"query"`
	ResponseMode     string                 `json:"response_mode"`
	ConversationID   string                 `json:"conversation_id,omitempty"`
	User             string                 `json:"user"`
	Files            []FileInput            `json:"files,omitempty"`
	AutoGenerateName *bool                  `json:"auto_generate_name,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatMessageResponse struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Mode           string `json:"mode"`
	Answer         string `json:"answer"`
	Metadata       struct {
		Usage Usage `json:"usage"`
	} `json:"metadata"`
	CreatedAt int64 `json:"created_at"`
}

func (api *API) ChatMessages(ctx context.Context, req *ChatMessageRequest) (resp *ChatMessageResponse, err error) {
	req.ResponseMode = "blocking"

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/chat-messages", req)
	if err != nil {
		return
	}
	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

func (api *API) StopChatMessage(ctx context.Context, taskID string, user string) error {
	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/chat-messages/"+taskID+"/stop", map[string]string{
		"user": user,
	})
	if err != nil {
		return err
	}
	return api.c.sendJSONRequest(httpReq, nil)
}
