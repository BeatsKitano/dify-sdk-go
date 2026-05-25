package dify

import (
	"context"
	"net/http"
)

type InfoResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        []any  `json:"tags"`
}

type SiteResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	IconType    string `json:"icon_type"`
	Copyright   string `json:"copyright"`
	Privacy     string `json:"privacy_policy"`
	ShowToken   bool   `json:"show_token"`
}

func (api *API) Info(ctx context.Context) (*InfoResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/info", nil)
	if err != nil {
		return nil, err
	}

	var resp InfoResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (api *API) Site(ctx context.Context) (*SiteResponse, error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, "/v1/site", nil)
	if err != nil {
		return nil, err
	}

	var resp SiteResponse
	if err := api.c.sendJSONRequest(httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
