package config

import (
	"reflect"
	"testing"
	"time"

	"exporter/internal/data"
)

func TestNewConfig(t *testing.T) {
	testConfig := NewConfig("../../example")
	want := Config{
		Server: Server{
			Path: "/metrics",
			Port: "8078",
			Url:  "https://monitoring.api.cloud.yandex.net/monitoring/v2/prometheusMetrics",
		},
		Instances: []Instance{
			{
				Token:             "<yandex token 1>",
				FolderId:          "<yandex folder id 1>",
				ServiceType:       "compute",
				SessionTimeoutSec: "5",
			},
			{
				Token:             "<yandex token 2>",
				FolderId:          "<yandex folder id 2>",
				ServiceType:       "compute",
				SessionTimeoutSec: "4",
			},
		},
	}

	if !reflect.DeepEqual(*testConfig, want) {
		t.Fatalf("ERROR: Not equal result '%v' and expected structure '%v'", testConfig, want)
	}
}

func TestConfig_InstanceRequestsData(t *testing.T) {
	tests := []struct {
		name                 string
		config               Config
		wantConfAfterDefault Config
		wantServerData       data.Server
		wantInstanceRequest  []data.InstanceRequest
	}{
		{
			name:   "Empty structures test",
			config: Config{},
			wantConfAfterDefault: Config{
				Server: Server{
					Path: "/metrics",
					Port: "8087",
					Url:  "https://monitoring.api.cloud.yandex.net/monitoring/v2/prometheusMetrics",
				},
				Instances: make([]Instance, 0),
			},
			wantServerData: data.Server{
				Path: "/metrics",
				Port: "8087",
			},
			wantInstanceRequest: make([]data.InstanceRequest, 0),
		},
		{
			name: "Setting all fields of structure test",
			config: Config{
				Server: Server{
					Path: "/test",
					Port: "1",
					Url:  "https://monitor.flant.com",
				},
				Instances: []Instance{
					{
						Token:             "Token 1",
						FolderId:          "Folder_Id_1",
						ServiceType:       "Custom_Service_Type_1",
						SessionTimeoutSec: "1",
					},
					{
						Token:             "Token 2",
						FolderId:          "Folder_Id_2",
						ServiceType:       "New_custom_Service_Type_2",
						SessionTimeoutSec: "2",
					},
				},
			},
			wantConfAfterDefault: Config{
				Server: Server{
					Path: "/test",
					Port: "1",
					Url:  "https://monitor.flant.com",
				},
				Instances: []Instance{
					{
						Token:             "Token 1",
						FolderId:          "Folder_Id_1",
						ServiceType:       "Custom_Service_Type_1",
						SessionTimeoutSec: "1",
					},
					{
						Token:             "Token 2",
						FolderId:          "Folder_Id_2",
						ServiceType:       "New_custom_Service_Type_2",
						SessionTimeoutSec: "2",
					},
				},
			},
			wantServerData: data.Server{
				Path: "/test",
				Port: "1",
			},
			wantInstanceRequest: []data.InstanceRequest{
				{
					Url:               "https://monitor.flant.com?folderId=Folder_Id_1&service=Custom_Service_Type_1",
					Token:             "Token 1",
					SessionTimeoutSec: 1 * time.Second,
				},
				{
					Url:               "https://monitor.flant.com?folderId=Folder_Id_2&service=New_custom_Service_Type_2",
					Token:             "Token 2",
					SessionTimeoutSec: 2 * time.Second,
				},
			},
		},
	}

	for _, tt := range tests {
		c := tt.config
		c.SetDefault()
		if !reflect.DeepEqual(c, tt.wantConfAfterDefault) {
			t.Fatalf("ERROR: Not equal config after preparing defaults values: '%v' and expected structure '%v'", c, tt.wantConfAfterDefault)
		}
		if !reflect.DeepEqual(c.ServerData(), tt.wantServerData) {
			t.Fatalf("ERROR: Not equal server data values: '%v' and expected structure '%v'", c.ServerData(), tt.wantServerData)
		}
		if !reflect.DeepEqual(c.InstanceRequestsData(), tt.wantInstanceRequest) {
			t.Fatalf("ERROR: Not equal server data values: '%v' and expected structure '%v'", c.InstanceRequestsData(), tt.wantInstanceRequest)
		}
	}
}
