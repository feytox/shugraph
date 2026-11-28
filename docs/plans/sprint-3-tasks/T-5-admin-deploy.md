Summary: infra: деплой graph-admin-frontapp

- Type: Task
- Priority: Major
- Assignee: feytox
- Tags: infra
- Parent: T-1
- Required for: GSD-37
- Soft Deadline: 09.12

~~~
## Описание

graph-admin-frontapp поднимается в `compose.yaml` и выкатывается на prod и stage-площадки отдельным релизом. Саму страницу запуска сценариев делает GSD-37.

## Решение

- [ ] graph-admin-frontapp поднимается в `compose.yaml`, его запросы к graph-admin-api проходят
- [ ] Workflow `deploy` собирает образ graph-admin-frontapp и выкатывает его отдельным релизом
- [ ] Снаружи админка открывается по своему префиксу пути, потому что `/` занят graph-frontapp
~~~
