package delivery

import (
	"github/Antihoman/Internet-proxy-server/cmd/internal/domain"
)

type Repository interface {
	GetAll() ([]domain.HTTPTransaction, error)
	Add(domain.HTTPTransaction) error
}