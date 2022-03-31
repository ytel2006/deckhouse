---
title: "Yandex cloud monitoring module: usage"
type:
  - instruction
search: Yandex cloud monitoring service
---

## An example of the module configuration

```yaml
monitoringYandexCloud: |
  apiKey: AAAAA
```

## Scrape metrics for the Yandex Cloud service

The list of available services is [here](https://cloud.yandex.ru/docs/monitoring/operations/metric/prometheusExport)

For example, we want to scrape metrics for the `interconnect` service.

Create a ServiceMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: yandex-interconnect-metrics
  namespace: d8-monitoring
spec:
  jobLabel: yandex-nat-instance
  endpoints:
    - port: https-metrics
      path: /metrics/interconnect # set name of service after /metrics/
      scheme: https
      bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
      tlsConfig:
        insecureSkipVerify: true
      honorLabels: true
      scrapeTimeout: "20s"
      # use the relabel mechanism if necessary
      # https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config
      # metricRelabelings:
      #  - targetLabel: "yandex_interconnect"
      #    replacement: "true"
  selector:
    matchLabels:
      app: yandex-cloud-metrics-exporter
  namespaceSelector:
    matchNames:
      - d8-monitoring
```

After creating the ServiceMonitor, the metrics will start scraping.
