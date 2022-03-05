package yandex

import (
	"fmt"

	dto "github.com/prometheus/client_model/go"
)

type natService struct {
	prefix       string
	instancesIds map[string]struct{}
}

func newNatInstanceService(cnf *Config) *natService {
	prefix := "nat_instance"
	if cnf.Prefix != "" {
		prefix = fmt.Sprintf("%s_%s", cnf.Prefix, prefix)
	}

	instancesIds := make(map[string]struct{})
	for _, id := range cnf.NatInstanceResourceIds {
		instancesIds[id] = struct{}{}
	}

	return &natService{
		prefix:       prefix,
		instancesIds: instancesIds,
	}
}

func (s *natService) IdForRequest() string {
	return natInstanceServiceId
}

func (s *natService) GetFilterFunc() func(metric *dto.Metric) *dto.Metric {
	return s.filter
}

func (s *natService) Prefix() string {
	return s.prefix
}

func (s *natService) filter(metric *dto.Metric) *dto.Metric {
	for _, l := range metric.Label {
		if l.GetName() != "resource_id" {
			continue
		}

		id := l.GetValue()

		if _, ok := s.instancesIds[id]; ok {
			return metric
		}

		break
	}

	return nil
}
