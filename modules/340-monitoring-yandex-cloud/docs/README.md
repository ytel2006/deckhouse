---
title: "The monitoring-yandex-cloud module"
---

The module is intended for scrape metrics from [Yandex.Cloud](https://cloud.yandex.ru/docs/monitoring/api-ref/MetricsData/prometheusMetrics).

When using the `WithNatInstance` placement scheme, the module automatically starts collecting metrics for the NAT host.
NAT-instance metrics will have `nat_instance=true` label. Also, module will deploy Grafana dashboard for NAT-instance metrics.

Before usage, please see [pricing policy for Yandex Cloud Monitoring](https://cloud.yandex.ru/docs/monitoring/pricing)
