package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int      `json:"limit,omitempty" validate:"gte=1,lte=20"`
	Offset int      `json:"offset,omitempty" validate:"gte=0"`
	Sort   string   `json:"sort,omitempty" validate:"oneof=asc desc"`
	Tags   []string `json:"tags,omitempty" validate:"max=5"`
	Search string   `json:"search,omitempty" validate:"max=100"`
	Since  string   `json:"since,omitempty"`
	Until  string   `json:"until,omitempty"`
}

func (pq PaginatedFeedQuery) Parse(req *http.Request) (PaginatedFeedQuery, error) {
	qs := req.URL.Query()
	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return pq, nil
		}
		pq.Limit = l
	}
	offset := qs.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return pq, nil
		}
		pq.Offset = o
	}
	sort := qs.Get("sort")
	if sort != "" {
		pq.Sort = sort
	}
	tags := qs.Get("tags")
	if tags != "" {
		pq.Tags = strings.Split(tags, ",")
	}
	search := qs.Get("search")
	if search != "" {
		pq.Search = search
	}
	since := qs.Get("since")
	if since != "" {
		pq.Since = ParseTime(since)
	}
	until := qs.Get("until")
	if until != "" {
		pq.Until = ParseTime(until)
	}
	return pq, nil
}

func ParseTime(str_time string) string {
	t, err := time.Parse(time.DateTime, str_time)
	if err != nil {
		return ""
	}

	return t.Format(time.DateTime)
}
