package httpx

type Meta struct {
	RequestID  string      `json:"request_id,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Limit   int   `json:"limit"`
	Offset  int   `json:"offset"`
	Total   int64 `json:"total"`
	HasMore bool  `json:"has_more"`
}

func NewPagination(limit, offset int, total int64) *Pagination {
	return &Pagination{
		Limit:   limit,
		Offset:  offset,
		Total:   total,
		HasMore: int64(offset+limit) < total,
	}
}
