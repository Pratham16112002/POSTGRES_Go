package paginate

import (
	"net/http"
	"strings"
)

type PostPaginateQuery struct {
	PaginatedQuery
	Sort string   `json:"sort,omitempty" validate:"oneof=asc desc"`
	Tags []string `json:"tags,omitempty" validate:"max=5"`
}

func (pq *PostPaginateQuery) Parse(req *http.Request) error {

	if err := pq.PaginatedQuery.Parse(req); err != nil {
		return err
	}
	pq.SetDefaults()
	qs := req.URL.Query()

	if sort := qs.Get("sort"); sort != "" {
		pq.Sort = sort
	}
	if tags := qs.Get("tags"); tags != "" {
		pq.Tags = strings.Split(tags, ",")
	}
	return nil
}

func (pq *PostPaginateQuery) SetDefaults() {
	pq.Sort = "asc"
}
