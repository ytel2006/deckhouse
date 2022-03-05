# k8s-yandex-exporter

Проект забирает метрики через Яндекс АПИ в готовом для prometheus формате. 
И добавляет метрики в prometheus, при помощи deckhouse.

Принцип получения метрик взят из [статьи](https://cloud.yandex.ru/docs/monitoring/operations/metric/prometheusExport) яндекс.
Сам экспортер не накапливает метрики. В ответ на запрос, он ходит через Яндекс Мониторинг АПИ, за значениями и отдает их в prometheus формате в виде странички.
Которую читает модуль deckhouse.

# Подготовка к работе с экспортером
1. Завести сервисный аккаунт в Яндекс облаке c ролью `monitoring.viewer`
2. Для аккаунта создать API-ключ для упрощенной аутентификации.
3. Сохранить полученное значение ключа в `.helm/secret-values.yaml` в поле `.global.api.token.`
4. Настроить конфиг экспортера в `.helm/templates/configMaps.yaml`
Здесь необходимо настроить поля:
```    instances:
      - token: {{ .Values.global.api.token }}
        folderId: "b1grXXXXXXref6c213ts"
        serviceType: "compute"
        sessionTimeoutSec: 5
```
Где поле `instances` - это массив, элементы которого будут различаться по `serviceType`.
В Яндекс АПИ метрики отдаются для всех инстансов одного типа в одном запросе.
Поэтому,чтобы различать метрику для конкретного инстанса, используются labels.

Пример полученной метрики:
```
# TYPE network_received_bytes gauge
network_received_bytes{interface_number="0", resource_id="instance-test-metrics", resource_type="vm"} 4.2
```

Поле `serviceType` выбирается из [списка](https://cloud.yandex.ru/docs/monitoring/operations/metric/prometheusExport)

Поле `folderId` берется, как id каталога с инфраструктурой Яндекс Облака.
Его можно достать из строки браузера `https://console.cloud.yandex.ru/folders/<FOLDER_ID>?section=dashboard`

Поле `sessionTimeoutSec` в секундах отвечает за время жизни сессии по пролучению метрик.

5. Дополнительная [информация](https://cloud.yandex.ru/docs/monitoring/pricing) о правилах тарификации Яндекс Метрик.