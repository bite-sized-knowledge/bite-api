package search

import "github.com/bite-sized/bite-api/internal/recsys"

type Service struct {
	recsysClient *recsys.Client
}

func NewService(recsysClient *recsys.Client) *Service {
	return &Service{recsysClient: recsysClient}
}
