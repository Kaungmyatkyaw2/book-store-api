package data

import (
	"math"
	"strings"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

type Metadata struct {
	CurrentPage  int `json:"currentPage"`
	PageSize     int `json:"pageSize"`
	FirstPage    int `json:"firstPage"`
	LastPage     int `json:"lastPage"`
	TotalRecords int `json:"totalRecords"`
}

func ValidateFilter(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter: " + f.Sort)

}

func (f Filters) sortDirecton() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {

	lastpage := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	metadata := Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		TotalRecords: totalRecords,
	}

	if lastpage == 0 {
		metadata.LastPage = 1
	} else {
		metadata.LastPage = lastpage
	}

	return metadata
}
