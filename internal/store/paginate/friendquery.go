package paginate

import "net/http"

type FriendPaginateQuery struct {
	PaginatedQuery
	Role string `json:"role,omitempty" validate:"oneof=user moderator"`
}

func (fq *FriendPaginateQuery) Parse(req *http.Request) error {
	if err := fq.PaginatedQuery.Parse(req); err != nil {
		return err
	}
	fq.SetDefaults()
	qs := req.URL.Query()

	if role := qs.Get("role"); role != "" {
		fq.Role = role
	}
	return nil
}

func (fq *FriendPaginateQuery) SetDefaults() {
	fq.Role = "user"
}
