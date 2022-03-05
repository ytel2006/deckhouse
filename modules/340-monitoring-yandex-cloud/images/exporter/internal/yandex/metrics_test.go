package yandex

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetPrefixForMetric(t *testing.T) {
	const testMetrics = `# TYPE disk_write_bytes gauge
disk_write_bytes{device="csi-1853288503", resource_id="test1-d1ce59cf-86774-wrnn8", resource_type="vm"} 0.0
disk_write_bytes{device="csi-3589549612", resource_id="test2-aaaaaaaa-bbbbb-ccccc", resource_type="vm"} 0.0
disk_write_bytes{device="csi-2461324810", resource_id="test3-dddddddd-eeeee-fffff", resource_type="vm"} 94754.13333333333
# TYPE disk_read_bytes gauge
disk_read_bytes{device="csi-3531230627", resource_id="test3-dddddddd-eeeee-fffff", resource_type="vm"} 0.0
# TYPE cpu_utilization gauge
cpu_utilization{cpu_name="cpu_1", resource_id="test1-d1ce59cf-86774-wrnn8", resource_type="vm"} 18.474740036436476
cpu_utilization{cpu_name="cpu_2", resource_id="test1-d1ce59cf-86774-wrnn8", resource_type="vm"} 37.01499249428546
cpu_utilization{cpu_name="cpu_1", resource_id="test2-aaaaaaaa-bbbbb-ccccc", resource_type="vm"} 18.474740036436476
cpu_utilization{cpu_name="cpu_2", resource_id="test2-aaaaaaaa-bbbbb-ccccc", resource_type="vm"} 37.01499249428546
`
	const prefix = "my_prefix"

	cnf := &Config{
		Prefix: prefix,
		NatInstanceResourceIds: []string{
			"test3-dddddddd-eeeee-fffff",
			"test1-d1ce59cf-86774-wrnn8",
		},
	}

	cases := []struct {
		service       Service
		shouldRenamed int
		metricsToTest []string
		title         string
	}{
		{
			title:         "generalService should rename metrics",
			service:       newGeneralService("some-service", cnf),
			shouldRenamed: 11,
			metricsToTest: []string{
				prefix + `_disk_write_bytes{device="csi-1853288503",resource_id="test1-d1ce59cf-86774-wrnn8",resource_type="vm"}`,
				prefix + `_disk_write_bytes{device="csi-3589549612",resource_id="test2-aaaaaaaa-bbbbb-ccccc",resource_type="vm"}`,
				prefix + `_disk_write_bytes{device="csi-2461324810",resource_id="test3-dddddddd-eeeee-fffff",resource_type="vm"}`,

				prefix + `_disk_read_bytes{device="csi-3531230627",resource_id="test3-dddddddd-eeeee-fffff",resource_type="vm"}`,

				prefix + `_cpu_utilization{cpu_name="cpu_1",resource_id="test1-d1ce59cf-86774-wrnn8",resource_type="vm"}`,
				prefix + `_cpu_utilization{cpu_name="cpu_1",resource_id="test2-aaaaaaaa-bbbbb-ccccc",resource_type="vm"}`,

				prefix + `_cpu_utilization{cpu_name="cpu_2",resource_id="test1-d1ce59cf-86774-wrnn8",resource_type="vm"}`,
				prefix + `_cpu_utilization{cpu_name="cpu_2",resource_id="test2-aaaaaaaa-bbbbb-ccccc",resource_type="vm"}`,
			},
		},

		{
			title:         "natService should filter only nat services metrics",
			service:       newNatInstanceService(cnf),
			shouldRenamed: 8,
			metricsToTest: []string{
				prefix + `_nat_instance_disk_write_bytes{device="csi-1853288503",resource_id="test1-d1ce59cf-86774-wrnn8",resource_type="vm"}`,
				prefix + `_nat_instance_disk_write_bytes{device="csi-2461324810",resource_id="test3-dddddddd-eeeee-fffff",resource_type="vm"}`,

				prefix + `_nat_instance_disk_read_bytes{device="csi-3531230627",resource_id="test3-dddddddd-eeeee-fffff",resource_type="vm"}`,

				prefix + `_nat_instance_cpu_utilization{cpu_name="cpu_1",resource_id="test1-d1ce59cf-86774-wrnn8",resource_type="vm"}`,
				prefix + `_nat_instance_cpu_utilization{cpu_name="cpu_2",resource_id="test1-d1ce59cf-86774-wrnn8",resource_type="vm"}`,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			body := bytes.NewBuffer([]byte(testMetrics))

			out, err := rewriteMetricsName(body, c.service)
			require.NoError(t, err)

			outStr := string(out)

			renamed := strings.Count(outStr, prefix)
			require.Equal(t, renamed, c.shouldRenamed)

			for _, s := range c.metricsToTest {
				require.True(t, strings.Contains(outStr, s))
			}
		})
	}

}
