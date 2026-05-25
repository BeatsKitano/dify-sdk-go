package dify

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
)

// --- Dataset types ---

type DatasetRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Permission  string `json:"permission,omitempty"`
	Provider    string `json:"provider,omitempty"`
}

type DatasetResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Provider       string `json:"provider"`
	Permission     string `json:"permission"`
	IndexingStatus string `json:"indexing_status"`
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

type DatasetListRequest struct {
	Keyword string `json:"keyword,omitempty"`
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
}

type DatasetListResponse struct {
	Page    int               `json:"page"`
	Limit   int               `json:"limit"`
	Total   int               `json:"total"`
	HasMore bool              `json:"has_more"`
	Data    []DatasetResponse `json:"data"`
}

type DatasetUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// --- Document types ---

type ProcessRule struct {
	Mode  string `json:"mode"`
	Rules *struct {
		PreProcessingRules []PreProcessingRule `json:"pre_processing_rules,omitempty"`
		Segmentation       *Segmentation       `json:"segmentation,omitempty"`
	} `json:"rules,omitempty"`
}

type PreProcessingRule struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

type Segmentation struct {
	Separator string `json:"separator"`
	MaxTokens int    `json:"max_tokens"`
}

type RetrievalModel struct {
	SearchMethod    string   `json:"search_method"`
	RerankingEnable bool     `json:"reranking_enable,omitempty"`
	TopK            int      `json:"top_k,omitempty"`
	ScoreThreshold  *float64 `json:"score_threshold,omitempty"`
}

type CreateDocumentByTextRequest struct {
	Name           string          `json:"name"`
	Text           string          `json:"text"`
	DocForm        string          `json:"doc_form,omitempty"`
	DocLanguage    string          `json:"doc_language,omitempty"`
	ProcessRule    *ProcessRule    `json:"process_rule,omitempty"`
	RetrievalModel *RetrievalModel `json:"retrieval_model,omitempty"`
}

type CreateDocumentByFileData struct {
	Name           string          `json:"name"`
	Text           string          `json:"text,omitempty"`
	DocForm        string          `json:"doc_form,omitempty"`
	DocLanguage    string          `json:"doc_language,omitempty"`
	ProcessRule    *ProcessRule    `json:"process_rule,omitempty"`
	RetrievalModel *RetrievalModel `json:"retrieval_model,omitempty"`
}

type CreateDocumentResponse struct {
	Document struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Status    string `json:"status"`
		CreatedAt int64  `json:"created_at"`
	} `json:"document"`
	Batch string `json:"batch"`
}

type DocumentIndexingStatusResponse struct {
	Data []DocumentIndexingStatusData `json:"data"`
}

type DocumentIndexingStatusData struct {
	ID                   string `json:"id"`
	IndexingStatus       string `json:"indexing_status"`
	ProcessingStartedAt  int64  `json:"processing_started_at"`
	ParsingCompletedAt   int64  `json:"parsing_completed_at"`
	CleaningCompletedAt  int64  `json:"cleaning_completed_at"`
	SplittingCompletedAt int64  `json:"splitting_completed_at"`
	CompletedAt          int64  `json:"completed_at"`
	Error                string `json:"error"`
	Enabled              bool   `json:"enabled"`
	DisabledAt           int64  `json:"disabled_at"`
	DisabledBy           string `json:"disabled_by"`
	Archived             bool   `json:"archived"`
	DisplayStatus        string `json:"display_status"`
	WordCount            int    `json:"word_count"`
	Tokens               int    `json:"tokens"`
}

type DocumentStatusUpdateRequest struct {
	DocumentIDs []string `json:"document_ids"`
}

// --- Segment types ---

type SegmentRequest struct {
	Content  string   `json:"content,omitempty"`
	Answer   string   `json:"answer,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
	Enabled  *bool    `json:"enabled,omitempty"`
}

type SegmentCreateRequest struct {
	Segments []SegmentRequest `json:"segments"`
}

type SegmentResponse struct {
	ID              string   `json:"id"`
	DocumentID      string   `json:"document_id"`
	Content         string   `json:"content"`
	Answer          string   `json:"answer"`
	Keywords        []string `json:"keywords"`
	SegmentPosition int      `json:"segment_position"`
	Status          string   `json:"status"`
	Enabled         bool     `json:"enabled"`
	CreatedAt       int64    `json:"created_at"`
	UpdatedAt       int64    `json:"updated_at"`
}

type SegmentCreateResponse struct {
	Data []SegmentResponse `json:"data"`
}

type SegmentListRequest struct {
	Status string `json:"status,omitempty"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

type SegmentListResponse struct {
	Data    []SegmentResponse `json:"data"`
	Page    int               `json:"page"`
	Limit   int               `json:"limit"`
	Total   int               `json:"total"`
	HasMore bool              `json:"has_more"`
}

type SegmentUpdateRequest struct {
	Content  string   `json:"content"`
	Answer   string   `json:"answer,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
	Enabled  *bool    `json:"enabled,omitempty"`
}

// --- Retrieve types ---

type RetrieveRequest struct {
	Query          string          `json:"query"`
	RetrievalModel *RetrievalModel `json:"retrieval_model,omitempty"`
}

type RetrieveResponse struct {
	Records []RetrieveRecord `json:"records"`
}

type RetrieveRecord struct {
	Segment RetrieveSegment `json:"segment"`
	Score   float64         `json:"score"`
}

type RetrieveSegment struct {
	ID         string `json:"id"`
	Content    string `json:"content"`
	Position   int    `json:"position"`
	DocumentID string `json:"document_id"`
	DatasetID  string `json:"dataset_id"`
}

// --- Dataset CRUD ---

func (api *API) CreateDataset(ctx context.Context, req *DatasetRequest) (*DatasetResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/datasets", req)
	if err != nil {
		return nil, err
	}

	var resp DatasetResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) Datasets(ctx context.Context, req *DatasetListRequest) (*DatasetListResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/datasets", nil)
	if err != nil {
		return nil, err
	}

	query := httpReq.URL.Query()
	if req.Keyword != "" {
		query.Set("keyword", req.Keyword)
	}
	if req.Page > 0 {
		query.Set("page", strconv.Itoa(req.Page))
	}
	if req.Limit > 0 {
		query.Set("limit", strconv.Itoa(req.Limit))
	}
	httpReq.URL.RawQuery = query.Encode()

	var resp DatasetListResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) UpdateDataset(ctx context.Context, datasetID string, req *DatasetUpdateRequest) (*DatasetResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodPatch, "/v1/datasets/"+datasetID, req)
	if err != nil {
		return nil, err
	}

	var resp DatasetResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) DeleteDataset(ctx context.Context, datasetID string) error {
	httpReq, err := api.createBaseRequest(ctx, http.MethodDelete, "/v1/datasets/"+datasetID, nil)
	if err != nil {
		return err
	}
	return api.c.sendDeleteRequest(httpReq)
}

// --- Document CRUD ---

func (api *API) CreateDocumentByText(ctx context.Context, datasetID string, req *CreateDocumentByTextRequest) (*CreateDocumentResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/datasets/"+datasetID+"/document/create-by-text", req)
	if err != nil {
		return nil, err
	}

	var resp CreateDocumentResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) CreateDocumentByFile(ctx context.Context, datasetID string, dataJSON string, filePath string, fileName string) (*CreateDocumentResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	httpReq, err := api.createMultipartRequest(ctx, http.MethodPost, "/v1/datasets/"+datasetID+"/document/create-by-file",
		map[string]string{"data": dataJSON},
		"file", fileName, file)
	if err != nil {
		return nil, err
	}

	var resp CreateDocumentResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) GetDocumentIndexingStatus(ctx context.Context, datasetID string, batch string) (*DocumentIndexingStatusResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/datasets/"+datasetID+"/documents/"+batch+"/indexing-status", nil)
	if err != nil {
		return nil, err
	}

	var resp DocumentIndexingStatusResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) UpdateDocumentStatus(ctx context.Context, datasetID string, action string, documentIDs []string) error {
	httpReq, err := api.createBaseRequest(ctx, http.MethodPatch, "/v1/datasets/"+datasetID+"/documents/status/"+action,
		&DocumentStatusUpdateRequest{DocumentIDs: documentIDs})
	if err != nil {
		return err
	}
	return api.c.sendJSONRequest(httpReq, nil)
}

// --- Segment CRUD ---

func (api *API) CreateSegments(ctx context.Context, datasetID string, documentID string, req *SegmentCreateRequest) (*SegmentCreateResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/datasets/"+datasetID+"/documents/"+documentID+"/segments", req)
	if err != nil {
		return nil, err
	}

	var resp SegmentCreateResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) Segments(ctx context.Context, datasetID string, documentID string, req *SegmentListRequest) (*SegmentListResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/datasets/"+datasetID+"/documents/"+documentID+"/segments", nil)
	if err != nil {
		return nil, err
	}

	query := httpReq.URL.Query()
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

	var resp SegmentListResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) UpdateSegment(ctx context.Context, datasetID string, documentID string, segmentID string, req *SegmentUpdateRequest) (*SegmentResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/datasets/"+datasetID+"/documents/"+documentID+"/segments/"+segmentID, req)
	if err != nil {
		return nil, err
	}

	var resp SegmentResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) DeleteSegment(ctx context.Context, datasetID string, documentID string, segmentID string) error {
	httpReq, err := api.createBaseRequest(ctx, http.MethodDelete, "/v1/datasets/"+datasetID+"/documents/"+documentID+"/segments/"+segmentID, nil)
	if err != nil {
		return err
	}
	return api.c.sendDeleteRequest(httpReq)
}

// --- Retrieve ---

func (api *API) Retrieve(ctx context.Context, datasetID string, req *RetrieveRequest) (*RetrieveResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/datasets/"+datasetID+"/retrieve", req)
	if err != nil {
		return nil, err
	}

	var resp RetrieveResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// MarshalDocumentData marshals document creation data to a JSON string for multipart file uploads.
func MarshalDocumentData(data *CreateDocumentByFileData) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
