Summary: graph-api: схема "адрес"

- Type: Task
- Priority: Major
- Assignee: Khasashin
- QA Assignee: goshayjakov
- Tags: backend
- Parent: T-1
- Depends on: GSD-47
- Soft Deadline: 09.12

~~~
## Описание

graph-api находит группы схемы "Один адрес на толпу" по SRD-4 (`docs/schemes/shared-address.md`) и считает их риск по `docs/schemes/risk.md`. Схема работает через свой детектор из GSD-47.

## Решение

- [ ] Группа создаётся, когда набран минимум SRD-4, и дополняется новыми событиями, пока не заморожена (SRD-9, SRD-10)
- [ ] Риск и `features.address` считаются при каждом обновлении группы, пороги и баллы в конфиге graph-api
- [ ] Интеграционные тесты на сценариях схемы из `backend/graph-admin-api/scenarios/`: состав групп и риск совпадают с ожидаемым ответом сценария
- [ ] @BarminGeorge принял схему по тест-плану на stage-площадке PR
~~~
