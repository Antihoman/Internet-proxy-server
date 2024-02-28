package delivery

import (
	"github/Antihoman/Internet-proxy-server/pkg/domain"
)

type Repository interface {
	GetAll() ([]domain.HTTPTransaction, error)
	GetByID(string) (domain.HTTPTransaction, error)
}