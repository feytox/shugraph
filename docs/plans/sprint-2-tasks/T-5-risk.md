Summary: graph-api: риск группы

- Type: Task
- Priority: Major
- Assignee: feytox
- Parent: T-1
- Depends on: T-4

~~~
## Описание

Риск группы первой схемы по `docs/schemes/risk.md`: `risk_score`, `risk_parts` и уровень. Пересчитывается при каждом дополнении группы.

## Тест-кейсы

1. Сценарий схемы из T-2: `risk_score` и `risk_parts` совпадают с ожидаемым ответом.
2. Дополнение группы событием меняет риск так, как следует из таблиц `risk.md`.
~~~
