Summary: backend: метрики сервисов

- Type: Task
- Priority: Normal
- Assignee: mefodyvaulin
- Tags: backend
- Parent: T-1
- Relates to: T-18

~~~
## Описание

graph-api и graph-admin-api отдают метрики по OTLP через OpenTelemetry SDK, а в логах есть `trace_id`. Куда слать, задаёт конфиг, хранилище и Collector поднимает T-18. За основу можно взять свой код из ДЗ по DevOps, если он подходит под пункты ниже.

## Решение

- [ ] OpenTelemetry SDK настраивается в `internal/app` обоих сервисов. Адрес и протокол отправки из стандартных переменных `OTEL_EXPORTER_OTLP_*`, без адреса метрики не отправляются и сервис работает как раньше
- [ ] У метрик те же атрибуты сервиса, что у логов: `service.name`, `service.version`, `deployment.environment.name`
- [ ] Метрики HTTP по семантическим соглашениям OpenTelemetry через `otelhttp`: число запросов, ошибки, длительность по маршруту
- [ ] Доменные метрики graph-api: принятые события по типу и результату (`201`, `204`, `409`), созданные группы и время сборки по схеме, решения аналитика. graph-admin-api: запуски сценариев
- [ ] Временная модель cumulative
- [ ] В JSON-логах запроса есть `trace_id`
- [ ] Тест: доменные метрики растут после приёма события и решения

## Тест-кейсы

1. Локально с Collector из T-18 прогнать `device_farm`: в VictoriaMetrics видны HTTP-метрики graph-api и рост числа событий и групп фермы.
~~~
