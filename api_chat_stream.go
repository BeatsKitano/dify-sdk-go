package dify

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ChatMessageStreamResponse struct {
	Event          string `json:"event"`
	TaskID         string `json:"task_id"`
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Answer         string `json:"answer"`
	CreatedAt      int64  `json:"created_at"`
}

type ChatMessageEndResponse struct {
	Event          string `json:"event"`
	TaskID         string `json:"task_id"`
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Metadata       struct {
		Usage              Usage  `json:"usage"`
		RetrieverResources []any `json:"retriever_resources"`
	} `json:"metadata"`
	CreatedAt int64 `json:"created_at"`
}

type AgentThoughtResponse struct {
	Event          string `json:"event"`
	ID             string `json:"id"`
	TaskID         string `json:"task_id"`
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Position       int    `json:"position"`
	Thought        string `json:"thought"`
	Observation    string `json:"observation"`
	Tool           string `json:"tool"`
	ToolLabels     map[string]interface{} `json:"tool_labels"`
	ToolInput      string                 `json:"tool_input"`
	MessageFiles   []interface{}          `json:"message_files"`
	CreatedAt      int64                  `json:"created_at"`
}

type MessageFileResponse struct {
	Event          string `json:"event"`
	ID             string `json:"id"`
	Type           string `json:"type"`
	BelongsTo      string `json:"belongs_to"`
	URL            string `json:"url"`
	ConversationID string `json:"conversation_id"`
}

type ChatMessageStreamChannelResponse struct {
	Event string `json:"event"`

	// message / message_end fields
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Answer         string `json:"answer"`
	CreatedAt      int64  `json:"created_at"`
	Metadata       *struct {
		Usage Usage `json:"usage"`
	} `json:"metadata,omitempty"`

	// agent_thought fields
	Thought     string `json:"thought"`
	Observation string `json:"observation"`
	Tool        string `json:"tool"`
	ToolInput   string `json:"tool_input"`

	Err error `json:"-"`
}

func (api *API) ChatMessagesStreamRaw(ctx context.Context, req *ChatMessageRequest) (*http.Response, error) {
	req.ResponseMode = "streaming"

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/chat-messages", req)
	if err != nil {
		return nil, err
	}
	return api.c.sendRequest(httpReq)
}

func (api *API) ChatMessagesStream(ctx context.Context, req *ChatMessageRequest) (chan ChatMessageStreamChannelResponse, error) {
	httpResp, err := api.ChatMessagesStreamRaw(ctx, req)
	if err != nil {
		return nil, err
	}

	streamChannel := make(chan ChatMessageStreamChannelResponse)
	go api.chatMessagesStreamHandle(ctx, httpResp, streamChannel)
	return streamChannel, nil
}

func (api *API) chatMessagesStreamHandle(ctx context.Context, resp *http.Response, streamChannel chan ChatMessageStreamChannelResponse) {
	defer resp.Body.Close()
	defer close(streamChannel)

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err.Error() != "EOF" {
					streamChannel <- ChatMessageStreamChannelResponse{
						Err: fmt.Errorf("error reading line: %w", err),
					}
				}
				return
			}

			if !bytes.HasPrefix(line, []byte("data: ")) {
				continue
			}
			line = bytes.TrimPrefix(line, []byte("data: "))

			var eventType struct {
				Event string `json:"event"`
			}
			if err := json.Unmarshal(line, &eventType); err != nil {
				streamChannel <- ChatMessageStreamChannelResponse{
					Err: fmt.Errorf("error unmarshalling event: %w", err),
				}
				return
			}

			switch eventType.Event {
			case "message":
				var msg ChatMessageStreamResponse
				if err := json.Unmarshal(line, &msg); err != nil {
					continue
				}
				streamChannel <- ChatMessageStreamChannelResponse{
					Event:          msg.Event,
					MessageID:      msg.MessageID,
					ConversationID: msg.ConversationID,
					Answer:         msg.Answer,
					CreatedAt:      msg.CreatedAt,
				}

			case "message_end":
				var end ChatMessageEndResponse
				if err := json.Unmarshal(line, &end); err != nil {
					continue
				}
				streamChannel <- ChatMessageStreamChannelResponse{
					Event:          end.Event,
					MessageID:      end.MessageID,
					ConversationID: end.ConversationID,
					Metadata: &struct {
						Usage Usage `json:"usage"`
					}{Usage: end.Metadata.Usage},
					CreatedAt: end.CreatedAt,
				}
				return

			case "agent_thought":
				var thought AgentThoughtResponse
				if err := json.Unmarshal(line, &thought); err != nil {
					continue
				}
				streamChannel <- ChatMessageStreamChannelResponse{
					Event:          thought.Event,
					MessageID:      thought.MessageID,
					ConversationID: thought.ConversationID,
					Thought:        thought.Thought,
					Observation:    thought.Observation,
					Tool:           thought.Tool,
					ToolInput:      thought.ToolInput,
				}

			case "message_file":
				var fileResp MessageFileResponse
				if err := json.Unmarshal(line, &fileResp); err != nil {
					continue
				}
				streamChannel <- ChatMessageStreamChannelResponse{
					Event:          fileResp.Event,
					ConversationID: fileResp.ConversationID,
				}

			case "workflow_started", "node_started", "node_finished", "workflow_finished":
				streamChannel <- ChatMessageStreamChannelResponse{
					Event: eventType.Event,
				}

			case "error":
				streamChannel <- ChatMessageStreamChannelResponse{
					Err: errors.New("stream error: " + string(line)),
				}
				return

			case "ping":
				// keep-alive, ignore
				continue

			default:
				continue
			}
		}
	}
}
