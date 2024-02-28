package delivery

import (
	"github/Antihoman/Internet-proxy-server/cmd/internal/domain"
)

type Repository interface {
	GetAll() ([]domain.Request, error)
	Add(req domain.Request) error
}