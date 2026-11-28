Summary: backend: метрики сервисов

- Type: Task
- Priority: Normal
- Assignee: mefodyvaulin
- Tags: backend
- Parent: T-1
- Relates to: T-18

~~~
## Описание

graph-api и graph-admin-api отдают метрики по OTLP через OpenTelemetry SDK. За основу можно взять свой код из ДЗ по DevOps, если он подходит под пункты ниже.

## Решение

- [ ] Метрики HTTP по семантическим соглашениям OpenTelemetry: число запросов, ошибки, длительность
- [ ] Доменные метрики: принятые события, найденные группы по схемам, время сборки по схеме, решения аналитика
- [ ] Метрики cumulative, адрес отправки задаётся конфигом
- [ ] В JSON-логах есть `trace_id`
~~~
