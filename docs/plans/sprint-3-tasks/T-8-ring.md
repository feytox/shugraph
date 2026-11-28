Summary: graph-api: схема "кольцо"

- Type: Task
- Priority: Major
- Assignee: Khasashin
- QA Assignee: goshayjakov
- Tags: backend
- Parent: T-1
- Depends on: GSD-47
- Soft Deadline: 16.12

~~~
## Описание

graph-api находит группы схемы "Кольцо" по SRD-6 (`docs/schemes/ring.md`) и считает их риск по `docs/schemes/risk.md`. Схема работает через свой детектор из GSD-47.

## Решение

- [ ] Неоднозначности правила SRD-6 до кода решены с @BarminGeorge и дописаны в правило
- [ ] Группа создаётся, когда набран минимум SRD-6, и дополняется новыми событиями, пока не заморожена (SRD-9, SRD-10)
- [ ] Риск и `features.ring` считаются при каждом обновлении группы, пороги и баллы в конфиге graph-api
- [ ] Интеграционные тесты на сценариях схемы из `backend/graph-admin-api/scenarios/`: состав групп и риск совпадают с ожидаемым ответом сценария, одиночный круг (`ring_single_round`) группу не создаёт
- [ ] @BarminGeorge принял схему по тест-плану на stage-площадке PR
~~~
