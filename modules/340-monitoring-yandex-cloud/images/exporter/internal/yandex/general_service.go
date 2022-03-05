package yandex

import (
	dto "github.com/prometheus/client_model/go"
)

type generalService struct {
	prefix string
	id     string
}

func newGeneralService(id string, cnf *Config) *generalService {
	return &generalService{
		prefix: cnf.Prefix,
		id:     id,
	}
}

func (s *generalService) IdForRequest() string {
	return s.id
}

func (s *generalService) GetFilterFunc() func(metric *dto.Metric) *dto.Metric {
	return nil
}

func (s *generalService) Prefix() string {
	return s.prefix
}
