# Dify Go SDK

Go SDK for the Dify API, supporting all Dify application types and Knowledge Base operations.

## Install

```bash
go get github.com/KevinZhao/dify-sdk-go
```

## Quick Start

```go
package main

import (
	"context"
	"log"
	"strings"

	"github.com/KevinZhao/dify-sdk-go"
)

func main() {
	ctx := context.Background()
	client := dify.NewClient("https://your-dify-host", "app-your-api-key")

	ch, err := client.API().ChatMessagesStream(ctx, &dify.ChatMessageRequest{
		Query: "Hello!",
		User:  "user-123",
	})
	if err != nil {
		log.Fatal(err)
	}

	var strBuilder strings.Builder
	for r := range ch {
		if r.Err != nil {
			log.Fatal(r.Err)
		}
		strBuilder.WriteString(r.Answer)
	}
	log.Println(strBuilder.String())
}
```

## Features

### Chat API
- `ChatMessages` / `ChatMessagesStream` - Send messages (blocking/streaming)
- `StopChatMessage` - Stop streaming generation
- `Messages` - Get conversation history
- `Conversations` - List conversations (supports `sort_by`)
- `ConversationsRenaming` - Rename conversation
- `DeleteConversation` - Delete conversation
- `MessagesFeedbacks` - Submit like/dislike feedback
- `SuggestedQuestions` - Get suggested follow-up questions
- `AppFeedbacks` - List all feedback for the app
- `Parameters` - Get application parameter info

### Completion API
- `CompletionMessages` / `CompletionMessagesStream` - Text generation (blocking/streaming)
- `StopCompletionMessage` - Stop generation

### Workflow API
- `RunWorkflow` / `RunStreamWorkflow` / `RunStreamWorkflowWithHandler` - Execute workflows
- `GetWorkflowRun` - Get workflow execution result
- `WorkflowLogs` - Query workflow execution logs
- `StopWorkflowTask` - Stop workflow task

### File API
- `UploadFile` - Upload files for chat/workflow
- `FilePreview` - Download/preview uploaded files

### Knowledge Base API
- `CreateDataset` / `Datasets` / `UpdateDataset` / `DeleteDataset` - Dataset CRUD
- `CreateDocumentByText` / `CreateDocumentByFile` - Create documents
- `GetDocumentIndexingStatus` - Check indexing progress
- `UpdateDocumentStatus` - Enable/disable/archive documents
- `CreateSegments` / `Segments` / `UpdateSegment` / `DeleteSegment` - Chunk management
- `Retrieve` - Test retrieval against a dataset

### General
- `Info` - Get application basic info
- `Site` - Get webapp settings

## Advanced Usage

### Per-call API secret override

```go
resp, err := client.API().WithSecret("app-another-key").ChatMessages(ctx, req)
```

### Custom HTTP client

```go
client := dify.NewClientWithConfig(&dify.ClientConfig{
	Host:             "https://your-dify-host",
	DefaultAPISecret: "app-your-api-key",
	Timeout:          30 * time.Second,
})
```

### Streaming workflow with full event handling

```go
handler := &dify.DefaultEventHandler{
	StreamHandler: func(resp dify.StreamingResponse) {
		switch resp.Event {
		case dify.EventWorkflowStarted:
			log.Println("Workflow started:", resp.WorkflowRunID)
		case dify.EventWorkflowFinished:
			log.Println("Workflow finished:", resp.Data.Outputs)
		}
	},
	TTSHandler: func(msg dify.TTSMessage) {
		log.Println("TTS audio length:", len(msg.Audio))
	},
	TextChunkHandler: func(msg dify.TextChunkResponse) {
		log.Println("Text chunk:", msg.Text)
	},
}

client.API().RunStreamWorkflowWithHandler(ctx, workflowReq, handler)
```

### Knowledge Base usage

```go
// Create a dataset
ds, _ := client.API().CreateDataset(ctx, &dify.DatasetRequest{
	Name: "My Knowledge Base",
})

// Add a document
doc, _ := client.API().CreateDocumentByText(ctx, ds.ID, &dify.CreateDocumentByTextRequest{
	Name: "Document 1",
	Text: "Document content here...",
	ProcessRule: &dify.ProcessRule{
		Mode: "automatic",
	},
})

// Check indexing status
status, _ := client.API().GetDocumentIndexingStatus(ctx, ds.ID, doc.Batch)

// Retrieve
results, _ := client.API().Retrieve(ctx, ds.ID, &dify.RetrieveRequest{
	Query: "search query",
})
```

## License

MIT
