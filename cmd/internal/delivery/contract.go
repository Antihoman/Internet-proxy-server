package delivery

import (
	"github/Antihoman/Internet-proxy-server/pkg/domain"
)

type Repository interface {
	Add(domain.HTTPTransaction) error
}