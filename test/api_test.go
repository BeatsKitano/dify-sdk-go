package test

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	dify "github.com/KevinZhao/dify-sdk-go"
)

var (
	host         = "YOUR_DIFY_HOST"
	apiSecretKey = "YOUR_API_SECRET_KEY"
)

func TestChatMessagesStreaming(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)
	ctx := context.Background()

	ch, err := client.API().ChatMessagesStream(ctx, &dify.ChatMessageRequest{
		Query: "Who are you?",
		User:  "test-user",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	var (
		strBuilder strings.Builder
		cId        string
	)
	for {
		select {
		case <-ctx.Done():
			t.Log("ctx.Done", strBuilder.String())
			return
		case r, isOpen := <-ch:
			if !isOpen {
				goto done
			}
			if r.Err != nil {
				t.Log("Stream error:", r.Err)
				return
			}
			strBuilder.WriteString(r.Answer)
			cId = r.ConversationID
			log.Println("Answer:", r.Answer, "ConversationID:", r.ConversationID, "MessageID:", r.MessageID)
		}
	}

done:
	t.Log(strBuilder.String())
	t.Log("Conversation ID:", cId)
}

func TestChatMessagesBlocking(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)
	ctx := context.Background()

	resp, err := client.API().ChatMessages(ctx, &dify.ChatMessageRequest{
		Query: "Hello",
		User:  "test-user",
	})
	if err != nil {
		t.Fatal(err)
	}
	j, _ := json.Marshal(resp)
	t.Log(string(j))
}

func TestMessages(t *testing.T) {
	cId := "ec373942-2d17-4f11-89bb-f9bbf863ebcc"
	ctx := context.Background()

	client := dify.NewClient(host, apiSecretKey)

	msg, err := client.API().Messages(ctx, &dify.MessagesRequest{
		ConversationID: cId,
		User:           "test-user",
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	j, _ := json.Marshal(msg)
	t.Log(string(j))
}

func TestMessagesFeedbacks(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)
	ctx := context.Background()

	res, err := client.API().MessagesFeedbacks(ctx, &dify.MessagesFeedbacksRequest{
		Rating: dify.FeedbackLike,
		User:   "test-user",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)
	log.Println(string(j))
}

func TestConversations(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)
	ctx := context.Background()

	res, err := client.API().Conversations(ctx, &dify.ConversationsRequest{
		User: "test-user",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)
	log.Println(string(j))
}

func TestConversationsRename(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)
	ctx := context.Background()

	res, err := client.API().ConversationsRenaming(ctx, &dify.ConversationsRenamingRequest{
		ConversationID: "ec373942-2d17-4f11-89bb-f9bbf863ebcc",
		Name:           "renamed-conversation",
		User:           "test-user",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)
	log.Println(string(j))
}

func TestParameters(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)
	ctx := context.Background()

	res, err := client.API().Parameters(ctx, &dify.ParametersRequest{
		User: "test-user",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)
	log.Println(string(j))
}

func TestRunWorkflow(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)

	workflowReq := dify.WorkflowRequest{
		Inputs: map[string]any{
			"image_url_new": map[string]string{
				"type":            "image",
				"transfer_method": "remote_url",
				"url":             "https://example.com/image.jpg",
			},
		},
		ResponseMode: "blocking",
		User:         "test-user",
	}

	resp, err := client.API().RunWorkflow(context.Background(), workflowReq)
	if err != nil {
		t.Fatalf("RunWorkflow error: %v", err)
	}

	if resp.WorkflowRunID == "" {
		t.Errorf("Expected non-empty WorkflowRunID")
	}
	if resp.Data.Status != "succeeded" {
		t.Errorf("Expected workflow status 'succeeded', got: %v", resp.Data.Status)
	}

	t.Logf("Workflow response: %+v", resp)
}

func TestRunWorkflowStreaming(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)

	workflowReq := dify.WorkflowRequest{
		Inputs: map[string]any{
			"image_url_new": map[string]string{
				"type":            "image",
				"transfer_method": "remote_url",
				"url":             "https://example.com/image.jpg",
			},
		},
		ResponseMode: "streaming",
		User:         "test-user",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var (
		mu               sync.Mutex
		workflowStarted  bool
		nodeStarted      bool
		nodeFinished     bool
		workflowFinished bool
	)

	handler := &testEventHandler{
		mu: &mu,
		onStreamingResponse: func(resp dify.StreamingResponse) {
			mu.Lock()
			defer mu.Unlock()
			switch resp.Event {
			case dify.EventWorkflowStarted:
				workflowStarted = true
			case dify.EventNodeStarted:
				nodeStarted = true
			case dify.EventNodeFinished:
				nodeFinished = true
			case dify.EventWorkflowFinished:
				workflowFinished = true
			}
		},
		onTTSMessage: func(msg dify.TTSMessage) {
			mu.Lock()
			defer mu.Unlock()
			t.Log("TTS audio length:", len(msg.Audio))
		},
		onTextChunk: func(msg dify.TextChunkResponse) {
			mu.Lock()
			defer mu.Unlock()
			t.Log("Text chunk:", msg.Text)
		},
	}

	err := client.API().RunStreamWorkflowWithHandler(ctx, workflowReq, handler)
	if err != nil {
		t.Fatalf("RunStreamWorkflow error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if !workflowStarted {
		t.Error("Expected workflow_started event")
	}
	if !nodeStarted {
		t.Error("Expected node_started event")
	}
	if !nodeFinished {
		t.Error("Expected node_finished event")
	}
	if !workflowFinished {
		t.Error("Expected workflow_finished event")
	}

	t.Log("Streaming workflow test completed")
}

type testEventHandler struct {
	mu                  *sync.Mutex
	onStreamingResponse func(dify.StreamingResponse)
	onTTSMessage        func(dify.TTSMessage)
	onTextChunk         func(dify.TextChunkResponse)
}

func (h *testEventHandler) HandleStreamingResponse(resp dify.StreamingResponse) {
	if h.onStreamingResponse != nil {
		h.onStreamingResponse(resp)
	}
}

func (h *testEventHandler) HandleTTSMessage(msg dify.TTSMessage) {
	if h.onTTSMessage != nil {
		h.onTTSMessage(msg)
	}
}

func (h *testEventHandler) HandleTextChunk(msg dify.TextChunkResponse) {
	if h.onTextChunk != nil {
		h.onTextChunk(msg)
	}
}
