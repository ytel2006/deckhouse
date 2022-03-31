---
title: "Yandex Cloud мониторинг: примеры конфигурации"
type:
  - instruction
search: Yandex Cloud мониторинг сервиса
---

## Пример конфигурации модуля

```yaml
monitoringYandexCloud: |
  apiKey: AAAAA
```

## Сбор метрик для сервиса Yandex Cloud

Список доступных сервисов находится [здесь](https://cloud.yandex.ru/docs/monitoring/operations/metric/prometheusExport)

Допустим, мы хотим собирать метрики для сервиса `interconnect`.

Создайте ServiceMonitor:   
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
      path: /metrics/interconnect # название сервиса указываем после /metrics/
      scheme: https
      bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
      tlsConfig:
        insecureSkipVerify: true
      honorLabels: true
      scrapeTimeout: "20s"
      # при необходимости воспользуйтесь механизмом relabel
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

После создания ServiceMonitor'а метрики начнут собираться.
