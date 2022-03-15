package yandex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	ycsdk "github.com/yandex-cloud/go-sdk"

	"github.com/pkg/errors"
	"github.com/yandex-cloud/go-sdk/iamkey"
)

const (
	prometheusMetricsUrl = "https://monitoring.api.cloud.yandex.net/monitoring/v2/prometheusMetrics"
	retries              = 3
)

type CloudApi struct {
	folderId        string
	logger          *log.Logger
	autoRenewPeriod time.Duration
	onRenewError    func()

	tokenMutex sync.RWMutex
	token      string

	isInit bool
	iamKey *iamkey.Key
}

func NewCloudApi(logger *log.Logger, folderId string) *CloudApi {
	return &CloudApi{
		folderId: folderId,
		logger:   logger,
		// iam token available during 12 hours, but yandex recommend update renew token every one hour
		autoRenewPeriod: 1 * time.Hour,
	}
}

func (a *CloudApi) WithAutoRenewPeriod(autoRenewPeriod time.Duration) *CloudApi {
	a.autoRenewPeriod = autoRenewPeriod

	return a
}

func (a *CloudApi) WithRenewTokenErrorHandler(handler func()) *CloudApi {
	a.onRenewError = handler

	return a
}

func (a *CloudApi) Init(serviceAccount io.Reader) error {
	if a.isInit {
		a.logger.Warningln("Yandex cloud api already init")
		return nil
	}

	var iamKey iamkey.Key
	decoder := json.NewDecoder(serviceAccount)

	err := decoder.Decode(&iamKey)
	if err != nil {
		return errors.Wrap(err, "malformed service account json")
	}

	a.iamKey = &iamKey

	err = a.renewToken()
	if err != nil {
		return err
	}

	go a.startAutoRenewToken()

	return nil
}

func (a *CloudApi) RequestMetrics(ctx context.Context, serviceId string) (io.ReadCloser, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 30 * time.Second,
			}).DialContext,
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.url(serviceId), nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating request: %s", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.token)

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

func (a *CloudApi) url(serviceId string) string {
	u, _ := url.Parse(prometheusMetricsUrl)

	query := u.Query()
	query.Set("service", serviceId)
	query.Set("folderId", a.folderId)

	u.RawQuery = query.Encode()

	return u.String()
}

func (a *CloudApi) getToken() string {
	a.tokenMutex.RLock()
	defer a.tokenMutex.RUnlock()

	return a.token
}

func (a *CloudApi) startAutoRenewToken() {
	a.logger.Info("Start auto renew IAM-token")
	a.logger.Warn("Stop auto renew IAM-token")

	t := time.NewTicker(a.autoRenewPeriod)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			err := a.renewToken()
			if err != nil {
				a.logger.Errorf("Cannot auto-renew IAM-token: %v", err)
				if a.onRenewError != nil {
					a.onRenewError()
				}
				return
			}
		}
	}
}

func (a *CloudApi) renewToken() error {
	token := ""

	rawCreds, err := ycsdk.ServiceAccountKey(a.iamKey)
	if err != nil {
		return errors.Wrap(err, "invalid auth credentials")
	}

	iamCreds, ok := rawCreds.(ycsdk.IAMTokenCredentials)
	if !ok {
		return fmt.Errorf("cannot convert to IAM-token")
	}

	var lastErr error
	for i := 1; i <= retries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cancel()
		resp, err := iamCreds.IAMToken(ctx)
		if err != nil {
			lastErr = errors.Wrap(err, "cannot get IAM-token")
			a.logger.Errorf("%v", lastErr)
			continue
		}

		token = resp.GetIamToken()
	}

	if token == "" {
		return fmt.Errorf("cannot get IAM-token after %d retries, last error: %v", retries, lastErr)
	}

	a.tokenMutex.Lock()
	defer a.tokenMutex.Unlock()
	a.token = token

	return nil
}
