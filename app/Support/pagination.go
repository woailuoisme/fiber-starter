package support

import (
	i18n "fiber-starter/app/Providers/I18n"

	"github.com/gofiber/fiber/v3"
)

// Paginator represents a paginated result set
type Paginator struct {
	Items interface{}    `json:"items"`
	Meta  PaginationMeta `json:"meta"`
}

// NewPaginator creates a new paginator instance from raw data
func NewPaginator(items interface{}, total int64, page, perPage int) *Paginator {
	return &Paginator{
		Items: items,
		Meta:  NewPaginationMeta(total, page, perPage),
	}
}

// Response returns a paginated fiber response
func (p *Paginator) Response(ctx fiber.Ctx, message ...string) error {
	msg := "Success"
	if len(message) > 0 {
		msg = message[0]
	}

	translated := i18n.Trans(ctx, msg, nil)
	if translated != msg {
		msg = translated
	}

	return HandleSuccess(ctx, msg, p)
}
