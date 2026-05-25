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

type CompletionMessageRequest struct {
	Inputs         map[string]any `json:"inputs"`
	Query          string         `json:"query"`
	ResponseMode   string         `json:"response_mode"`
	User           string         `json:"user"`
	Files          []FileInput    `json:"files,omitempty"`
}

type CompletionMessageResponse struct {
	MessageID string `json:"message_id"`
	Mode      string `json:"mode"`
	Answer    string `json:"answer"`
	Metadata  struct {
		Usage Usage `json:"usage"`
	} `json:"metadata"`
	CreatedAt int64 `json:"created_at"`
}

type CompletionStreamResponse struct {
	Event     string `json:"event"`
	MessageID string `json:"message_id"`
	Answer    string `json:"answer"`
	CreatedAt int64  `json:"created_at"`
}

type CompletionStreamChannelResponse struct {
	Event     string `json:"event"`
	MessageID string `json:"message_id"`
	Answer    string `json:"answer"`
	CreatedAt int64  `json:"created_at"`
	Audio     string `json:"audio"`
	Metadata  *struct {
		Usage Usage `json:"usage"`
	} `json:"metadata,omitempty"`
	Err error `json:"-"`
}

func (api *API) CompletionMessages(ctx context.Context, req *CompletionMessageRequest) (*CompletionMessageResponse, error) {
	req.ResponseMode = "blocking"

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/completion-messages", req)
	if err != nil {
		return nil, err
	}

	var resp CompletionMessageResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) CompletionMessagesStream(ctx context.Context, req *CompletionMessageRequest) (chan CompletionStreamChannelResponse, error) {
	req.ResponseMode = "streaming"

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/completion-messages", req)
	if err != nil {
		return nil, err
	}

	resp, err := api.c.sendRequest(httpReq)
	if err != nil {
		return nil, err
	}

	streamChannel := make(chan CompletionStreamChannelResponse)
	go api.completionMessagesStreamHandle(ctx, resp, streamChannel)
	return streamChannel, nil
}

func (api *API) completionMessagesStreamHandle(ctx context.Context, resp *http.Response, streamChannel chan CompletionStreamChannelResponse) {
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
				streamChannel <- CompletionStreamChannelResponse{
					Err: fmt.Errorf("error unmarshalling event: %w", err),
				}
				return
			}

			switch eventType.Event {
			case "message":
				var msg CompletionStreamResponse
				if err := json.Unmarshal(line, &msg); err != nil {
					continue
				}
				streamChannel <- CompletionStreamChannelResponse{
					Event:     msg.Event,
					MessageID: msg.MessageID,
					Answer:    msg.Answer,
					CreatedAt: msg.CreatedAt,
				}

			case "message_end":
				var end struct {
					Event     string `json:"event"`
					MessageID string `json:"message_id"`
					Metadata  struct {
						Usage Usage `json:"usage"`
					} `json:"metadata"`
					CreatedAt int64 `json:"created_at"`
				}
				if err := json.Unmarshal(line, &end); err != nil {
					continue
				}
				streamChannel <- CompletionStreamChannelResponse{
					Event:     end.Event,
					MessageID: end.MessageID,
					CreatedAt: end.CreatedAt,
					Metadata: &struct {
						Usage Usage `json:"usage"`
					}{Usage: end.Metadata.Usage},
				}
				return

			case "tts_message":
				var tts struct {
					Event     string `json:"event"`
					MessageID string `json:"message_id"`
					Audio     string `json:"audio"`
					CreatedAt int64  `json:"created_at"`
				}
				if err := json.Unmarshal(line, &tts); err != nil {
					continue
				}
				streamChannel <- CompletionStreamChannelResponse{
					Event:     tts.Event,
					MessageID: tts.MessageID,
					Audio:     tts.Audio,
					CreatedAt: tts.CreatedAt,
				}

			case "tts_message_end":
				var ttsEnd struct {
					Event     string `json:"event"`
					MessageID string `json:"message_id"`
					CreatedAt int64  `json:"created_at"`
				}
				if err := json.Unmarshal(line, &ttsEnd); err != nil {
					continue
				}
				streamChannel <- CompletionStreamChannelResponse{
					Event:     ttsEnd.Event,
					MessageID: ttsEnd.MessageID,
					CreatedAt: ttsEnd.CreatedAt,
				}

			case "error":
				streamChannel <- CompletionStreamChannelResponse{
					Err: errors.New("stream error: " + string(line)),
				}
				return

			case "ping":
				continue

			default:
				continue
			}
		}
	}
}

func (api *API) StopCompletionMessage(ctx context.Context, taskID string, user string) error {
	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/completion-messages/"+taskID+"/stop", map[string]string{
		"user": user,
	})
	if err != nil {
		return err
	}
	return api.c.sendJSONRequest(httpReq, nil)
}
