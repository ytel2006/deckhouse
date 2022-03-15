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
	"time"

	ycsdk "github.com/yandex-cloud/go-sdk"

	"github.com/pkg/errors"
	"github.com/yandex-cloud/go-sdk/iamkey"
)

const prometheusMetricsUrl = "https://monitoring.api.cloud.yandex.net/monitoring/v2/prometheusMetrics"

type CloudApi struct {
	token    string
	folderId string
}

func NewCloudApi(token, folderId string) *CloudApi {
	return &CloudApi{
		token:    token,
		folderId: folderId,
	}
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

func TokenFromServiceAccount(sa io.Reader) (string, error) {
	var iamKey iamkey.Key

	decoder := json.NewDecoder(sa)

	err := decoder.Decode(&iamKey)
	if err != nil {
		return "", errors.Wrap(err, "malformed service account json")
	}

	creds, err := ycsdk.ServiceAccountKey(&iamKey)

	c, ok := creds.(ycsdk.IAMTokenCredentials)
	if !ok {
		return "", fmt.Errorf("cannot convert to iam token")
	}

	if err != nil {
		return "", errors.Wrap(err, "invalid auth credentials")
	}

	resp, err := c.IAMToken(context.TODO())
	if err != nil {
		return "", err
	}

	// todo fix
	return resp.GetIamToken(), nil
}
