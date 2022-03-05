package yandex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

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

type Metrics struct {
	config *Config

	natInstanceService Service
}

type Service interface {
	Prefix() string
	Filter(m *dto.Metric) *dto.Metric
	HasFilter() bool
	IdForRequest() string
}

func NewMetrics(config *Config) *Metrics {
	return &Metrics{
		config: config,

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

	service := m.service(serviceId)

	bodyReader, err := m.request(ctx, service.IdForRequest())
	if err != err {
		return nil, fmt.Errorf("error do request for service %s: %v", serviceId, err)
	}

	defer bodyReader.Close()

	return rewriteMetricsName(bodyReader, service)
}

func (m *Metrics) service(inService string) Service {
	if inService == natInstanceServiceId {
		return m.natInstanceService
	}

	return newGeneralService(inService, m.config)
}

func rewriteMetricsName(body io.Reader, service Service) ([]byte, error) {
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

func filterMetrics(service Service, mf *dto.MetricFamily) {
	if !service.HasFilter() {
		return
	}

	metrics := mf.GetMetric()
	newMetrics := make([]*dto.Metric, 0, len(mf.GetMetric()))

	for _, m := range metrics {
		newMetric := service.Filter(m)
		if newMetric != nil {
			newMetrics = append(newMetrics, newMetric)
		}
	}

	mf.Metric = newMetrics
}

func (m *Metrics) request(ctx context.Context, serviceId string) (io.ReadCloser, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 30 * time.Second,
			}).DialContext,
		},
	}

	const url = "https://monitoring.api.cloud.yandex.net/monitoring/v2/prometheusMetrics"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating request: %s", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.config.Token)

	response, err := client.Do(req)

	if e, ok := err.(net.Error); ok && e.Timeout() {
		return nil, fmt.Errorf("do request timeout: %s", err)
	} else if err != nil {
		return nil, fmt.Errorf("failed send request: %s", err)
	}

	if response.StatusCode != http.StatusOK {
		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			errStr := fmt.Errorf("parse error for response body: %s", err).Error()
			responseData = []byte(errStr)
		}

		response.Body.Close()

		return nil, fmt.Errorf("status code %v, error response: %s", response.StatusCode, string(responseData))
	}

	return response.Body, nil
}
