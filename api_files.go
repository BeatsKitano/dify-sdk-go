package dify

import (
	"context"
	"net/http"
	"os"
)

type FileUploadResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Extension string `json:"extension"`
	MimeType  string `json:"mime_type"`
	CreatedBy int    `json:"created_by"`
	CreatedAt int64  `json:"created_at"`
}

func (api *API) UploadFile(ctx context.Context, filePath string, user string) (*FileUploadResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	httpReq, err := api.createMultipartRequest(ctx, http.MethodPost, "/v1/files/upload",
		map[string]string{"user": user},
		"file", fileInfo.Name(), file)
	if err != nil {
		return nil, err
	}

	var resp FileUploadResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) FilePreview(ctx context.Context, fileID string, asAttachment bool) (*http.Response, error) {
	url := "/v1/files/" + fileID + "/preview"
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if asAttachment {
		query := httpReq.URL.Query()
		query.Set("as_attachment", "true")
		httpReq.URL.RawQuery = query.Encode()
	}

	return api.c.sendRequest(httpReq)
}
