Summary: graph-api: схема "цепочка"

- Type: Task
- Priority: Major
- Assignee: mefodyvaulin
- QA Assignee: goshayjakov
- Tags: backend
- Parent: T-1
- Depends on: GSD-47
- Soft Deadline: 16.12

~~~
## Описание

graph-api находит группы схемы "Быстрая цепочка" по SRD-5 (`docs/schemes/fast-chain.md`) и считает их риск по `docs/schemes/risk.md`. Схема работает через свой детектор из GSD-47.

## Решение

- [ ] Неоднозначности правила SRD-5 до кода решены с @BarminGeorge и дописаны в правило
- [ ] Группа создаётся, когда набран минимум SRD-5, и дополняется новыми событиями, пока не заморожена (SRD-9, SRD-10)
- [ ] Звено цепочки - перевод со счёта, на который пришёл предыдущий перевод, и не раньше него. Путь ищется по переводам в окне SRD-1 рекурсивным CTE в Postgres или в Go, выбор и причина описаны в PR
- [ ] В группу входят только переводы найденного пути и счета на нём, `hops_count` и `preserved_percent` считаются так, как описано в `docs/cards/fast-chain.md`
- [ ] Риск и `features.chain` считаются при каждом обновлении группы, пороги и баллы в конфиге graph-api
- [ ] Интеграционные тесты на сценариях схемы из `backend/graph-admin-api/scenarios/`: состав групп и риск совпадают с ожидаемым ответом сценария
- [ ] @BarminGeorge принял схему по тест-плану на stage-площадке PR
~~~
