Summary: infra: логи и метрики prod в Grafana

- Type: Task
- Priority: Normal
- Assignee: feytox
- Tags: infra
- Parent: T-1

~~~
## Описание

Логи и метрики prod собираются в VictoriaLogs и VictoriaMetrics и видны в Grafana. Мониторинг живёт в своём namespace и от prod и stage не зависит. Метрики в коде сервисов делает T-21, дашборды T-22.

## Решение

- [ ] VictoriaLogs, VictoriaMetrics, OpenTelemetry Collector и Grafana в namespace `monitoring`, поды с приоритетом prod
- [ ] Collector читает логи подов prod и принимает OTLP от сервисов
- [ ] Метрики подов и узла собираются без правок в сервисах
- [ ] Grafana открывается по HTTPS на поддомене со своим входом
~~~
