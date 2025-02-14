---
title: "Модуль linstor: FAQ"
---

<div class="docs__information warning active">
Модуль находится в процессе активного развития, и его функциональность может существенно измениться.
</div>

## Когда следует использовать LVM, а когда LVMThin?

Если кратко, то:
- LVM проще и обладает производительностью сравнимой с производительностью накопителя;
- LVMThin позволяет использовать snapshot'ы и overprovisioning, но медленнее в два раза.

Подробнее — в следующем разделе.

## Производительность и надёжность LINSTOR, а также сравнение с Ceph

Мы придерживаемся практического взгляда на вопрос. Разница в несколько десятков процентов на практике никогда не имеет значения. Имеет значение разница в несколько раз и более.

Факторы сравнения:
- Последовательное чтение и запись: не имеют никакого значения, потому что на любой технологии они всегда упираются в сеть (что 10 Гбит/с, что 1 Гбит/с). С практической точки зрения этот показатель можно полностью игнорировать;
- Случайное чтение и запись (что на 1Гбит/с, что на 10Гбит/с):
  - drbd+lvm в 5 раз лучше ceph-rbd (latency — в 5 раз меньше, IOPS — в 5 раз больше);
  - drbd+lvm в 2 раза лучше drbd+lvmthin.
- Если одна из реплик расположена локально, то скорость чтения будет примерно равна скорости устройства хранения;
- Если нет реплик расположенных локально, то скорость записи будет примерно ограничена половиной пропускной способности сети при двух репликах или ⅓ пропускной способности сети при трех репликах;
- При большом количестве клиентов (больше 10, при iodepth 64), ceph начинает отставать сильнее (до 10 раз) и потреблять значительно больше CPU.

В сухом остатке получается, что на практике неважно какие параметры менять, и есть всего три значимых фактора:
- **Локальность чтения** — если всё чтение производится локально, то оно работает со скоростью (throughput, IOPS, latency) локального диска (разница практически незаметна); 
- **1 сетевой hop при записи** — в drbd репликацией занимается *клиент*, а в ceph — *сервер*, поэтому у ceph latency на запись всегда минимум в два раза больше чем у drbd;
- **Сложность кода** — latency вычислений на datapath (сколько процессорных команд выполняется на каждую операцию ввода/вывода), — drbd+lvm проще чем drbd+lvmthin, и значительно проще чем ceph-rbd.

## Что использовать в какой ситуации?

По умолчанию модуль использует две реплики (третья — для кворума, diskless, создается автоматически). Такой подход гарантирует защиту от split-brain и достаточный уровень надежности хранения, но нужно учитывать следующие особенности:
  - В момент недоступности одной из реплик (реплика A) данные записываются только в единственную реплику (реплика B). Это означает, что:
    - Если в этот момент отключится и вторая реплика (реплика B), то запись и чтение будут недоступны;
    - Если при этом вторая реплика (реплика B) утеряна безвозвратно, то данные будут частично потеряны (есть только старая реплика A);
    - Если старая реплика (реплика A) была тоже утеряна безвозвратно, то данные будут потеряны полностью.
  - Чтобы включиться обратно при отключении второй реплики (без вмешательства оператора) требуется доступность обеих реплик. Это необходимо, чтобы корректно отработать ситуацию split-brain;
  - Включение третьей реплики решает обе проблемы (в любой момент времени доступно минимум две копии данных), но увеличивает накладные расходы (сеть, диск).

Настоятельно рекомендуется иметь одну реплику локально. Это в два раза увеличивает возможную скорость запись (при двух репликах) и значительно увеличивает скорость чтения. Но даже если реплики на локальном хранилище нет, то все также будет работать нормально, за исключением того, что чтение будет осуществляться по сети и будет двойная утилизация сети при записи.

В зависимости от задачи нужно выбрать один из следующих вариантов:
- drbd+lvm — быстрей (в два раза) и надежней (lvm — проще);
- drbd+lvmthin — поддержка snapshot'ов и возможность overprovisioning.

## Как добавить существующий LVM или LVMThin-пул?

Пример добавления LVM-пула:
```shell
linstor storage-pool create lvm node01 lvmthin linstor_data
```

Пример добавления LVMThin-пула:
```shell
linstor storage-pool create lvmthin node01 lvmthin linstor_data/data
```

Можно добавлять и пулы, в которых уже созданы какие-то тома. LINSTOR просто будет создавать в пуле новые тома.

## Как настроить Prometheus на использование хранилища LINSTOR?

Чтобы настроить Prometheus на использование хранилища LINSTOR, необходимо:
- Настроить пулы хранения и StorageClass, как показано в [примерах использования](usage.html);
- Указать параметры [longtermStorageClass](../300-prometheus/configuration.html#parameters-longtermstorageclass) и [storageClass](../300-prometheus/configuration.html#parameters-storageclass) в конфигурации модуля [prometheus](../300-prometheus/). Например:
  ```yaml
  prometheus: |
    longtermStorageClass: linstor-data-r2
    storageClass: linstor-data-r2
  ```
- Дождаться перезапуска Pod'ов Prometheus.

## Pod не может запуститься из-за ошибки `FailedMount`

### Pod завис на стадии `ContainerCreating`
Если Pod завис на стадии `ContainerCreating`, а в выводе `kubectl describe` есть ошибки вида:

```
rpc error: code = Internal desc = NodePublishVolume failed for pvc-b3e51b8a-9733-4d9a-bf34-84e0fee3168d: checking for exclusive open failed: wrong medium type, check device health
```

Значит устройство всё ещё смонтировано на одном из других узлов. Проверить это можно с помощью следующей команды:
```shell
linstor r l -r pvc-b3e51b8a-9733-4d9a-bf34-84e0fee3168d
```

Флаг `InUse` укажет на каком узле используется устройство.

### Ошибки вида `Input/output error`

Такие ошибки обычно возникают на стадии создания файловой системы (mkfs).

Проверьте `dmesg` на узле, где запускается Pod:
```shell
dmesg | grep 'Remote failed to finish a request within'
```

Если вывод команды не пустой (в выводе `dmesg` есть строки вида "Remote failed to finish a request within ..."), то скорее всего, ваша дисковая подсистема слишком медленная для нормального функционирования DRBD.
