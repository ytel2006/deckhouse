---
title: "Модуль monitoring-yandex-cloud"
---

Модуль предназначен для сбора метрик с [Yandex.Cloud](https://cloud.yandex.ru/docs/monitoring/api-ref/MetricsData/prometheusMetrics)

При использовании схемы размещения `WithNatInstance` модуль автоматически начинает собирать метрики для NAT-узла.
Метрики для NAT-узла будут иметь лейбл `nat_instance=true`. Также модуль создает дашборду для NAT-узла в Grafana.

Перед использованием, пожалуйста ознакомьтесь с [правилами тарификации Yandex Cloud Monitoring](https://cloud.yandex.ru/docs/monitoring/pricing) 
