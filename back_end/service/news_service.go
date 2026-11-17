package service

import (
	"backend/model/model"
	"backend/utils"
)

type NewsService interface {
	GetCryptoNews(symbol string) ([]model.NewsItem, error)
}

type newsServiceImpl struct{}

func NewNewsService() NewsService {
	return &newsServiceImpl{}
}

func (s *newsServiceImpl) GetCryptoNews(symbol string) ([]model.NewsItem, error) {
	return utils.FetchAndAnalyzeCryptoNews(symbol)
}
