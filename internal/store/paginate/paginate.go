package paginate

import (
	"net/http"
	"strconv"
)

type PaginatedQuery struct {
	Limit  int    `json:"limit,omitempty" validate:"gte=1,lte=20"`
	Offset int    `json:"offset,omitempty" validate:"gte=0"`
	Search string `json:"search,omitempty" validate:"max=100"`
}

func (pq *PaginatedQuery) Parse(req *http.Request) error {
	qs := req.URL.Query()

	// Parse Limit
	if limit := qs.Get("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return err
		}
		pq.Limit = l
	}

	// Parse Offset
	if offset := qs.Get("offset"); offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return err
		}
		pq.Offset = o
	}

	// Parse Search
	if search := qs.Get("search"); search != "" {
		pq.Search = search
	}

	return nil
}

func (pq *PaginatedQuery) SetDefaults() {
	pq.Limit = 20
	pq.Offset = 0
	pq.Search = ""
}
