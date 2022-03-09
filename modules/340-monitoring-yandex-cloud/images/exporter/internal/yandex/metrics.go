package yandex

import (
	"bytes"
	"context"
	"fmt"
	"io"

	dto "github.com/prometheus/client_model/go"

	"github.com/prometheus/common/expfmt"
)

const (
	natInstanceServiceId = "nat-instance"
	computeService       = "compute"
)

var services = map[string]struct{}{
	// Compute Cloud
	natInstanceServiceId: {},
	// Compute Cloud
	computeService: {},
	// Object Storage
	"storage": {},
	// Managed Service for PostgreSQL
	"managed-postgresql": {},
	// Managed Service for ClickHouse;
	"managed-clickhouse": {},
	// Managed Service for MongoDB;
	"managed-mongodb": {},
	// Managed Service for MySQL;
	"managed-mysql": {},
	// Managed Service for Redis;
	"managed-redis": {},
	// Managed Service for Apache Kafka®;
	"managed-kafka": {},
	// Managed Service for Elasticsearch;
	"managed-elasticsearch": {},
	// Managed Service for SQL Server
	"managed-sqlserver": {},
	// Managed Service for Kubernetes;
	"managed-kubernetes": {},
	// Cloud Functions
	"serverless-functions": {},
	// триггеры Cloud Functions
	"serverless_triggers_client_metrics": {},
	// Yandex Database
	"ydb": {},
	// Cloud Interconnect;
	"interconnect": {},
	// Certificate Manager;
	"certificate-manager": {},
	// Data Transfer
	"data-transfer": {},
	// Data Proc
	"data-proc": {},
	// API Gateway.
	"serverless-apigateway": {},
}

type Config struct {
	Token    string `yaml:"token"`
	FolderId string `yaml:"folderId"`
	Prefix   string `yaml:"prefix"`

	NatInstanceResourceIds []string `yaml:"natInstanceResourceIds"`
}

type service interface {
	Prefix() string
	GetFilterFunc() func(m *dto.Metric) *dto.Metric
	IdForRequest() string
}

type Api interface {
	RequestMetrics(ctx context.Context, serviceId string) (io.ReadCloser, error)
}

type Metrics struct {
	config *Config
	api    Api

	natInstanceService service
}

func NewMetrics(config *Config, api Api) *Metrics {
	return &Metrics{
		config: config,
		api:    api,

		natInstanceService: newNatInstanceService(config),
	}
}

func (m *Metrics) HasService(service string) bool {
	_, ok := services[service]

	return ok
}

func (m *Metrics) GetMetrics(ctx context.Context, serviceId string) ([]byte, error) {
	if !m.HasService(serviceId) {
		return nil, fmt.Errorf("Service not exists: %s", serviceId)
	}

	svc := m.service(serviceId)

	bodyReader, err := m.api.RequestMetrics(ctx, svc.IdForRequest())
	if err != err {
		return nil, fmt.Errorf("error do request for service %s: %v", serviceId, err)
	}

	defer bodyReader.Close()

	return rewriteMetricsName(bodyReader, svc)
}

func (m *Metrics) service(inService string) service {
	if inService == natInstanceServiceId {
		return m.natInstanceService
	}

	return newGeneralService(inService, m.config)
}

func rewriteMetricsName(body io.Reader, service service) ([]byte, error) {
	var parser expfmt.TextParser
	families, err := parser.TextToMetricFamilies(body)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer

	prefix := service.Prefix()

	for name, mf := range families {
		if prefix != "" {
			newPrefix := fmt.Sprintf("%s_%s", prefix, mf.GetName())
			mf.Name = &newPrefix
		}
		filterMetrics(service, mf)
		_, err := expfmt.MetricFamilyToText(&out, mf)
		if err != nil {
			return nil, fmt.Errorf("Error while metric family to text %s: %v", name, err)
		}

	}

	return out.Bytes(), nil
}

func filterMetrics(service service, mf *dto.MetricFamily) {
	filter := service.GetFilterFunc()
	if filter == nil {
		return
	}

	metrics := mf.GetMetric()
	newMetrics := make([]*dto.Metric, 0, len(mf.GetMetric()))

	for _, m := range metrics {
		newMetric := filter(m)
		if newMetric != nil {
			newMetrics = append(newMetrics, newMetric)
		}
	}

	mf.Metric = newMetrics
}
