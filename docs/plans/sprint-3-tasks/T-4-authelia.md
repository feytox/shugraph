Summary: infra: вход через Authelia

- Type: Task
- Priority: Major
- Assignee: feytox
- Tags: infra
- Parent: T-1
- Depends on: T-2
- Required for: T-11, T-9
- Soft Deadline: 09.12

~~~
## Описание

Без логина на prod и stage-площадки не зайти: вход проверяет Authelia через ForwardAuth в Traefik, свою аутентификацию не пишем. Учётки заводим сами, регистрации нет. Админка и `/admin-api` открыты только группе админов.

## Решение

- [ ] В первый день спайк: middleware ForwardAuth цепляется к HTTPRoute через `ExtensionRef` на нашей версии Traefik. Если не выходит, обсуждаем запасной вариант до продолжения задачи
- [ ] Authelia поднимается в своём namespace, страница входа на `auth.shugraph.feytox.ru`, сессия общая для prod и всех `pr-<N>`
- [ ] Учётки и группы в файле пользователей Authelia, файл в k8s Secret: у каждого из команды учётка в группе админов, плюс учётки аналитика для показа и жюри
- [ ] Все маршруты prod и stage-площадок идут через ForwardAuth; админка, `/admin-api` и ручка очистки БД в graph-api (T-12) - только для группы админов
- [ ] Сервисы получают `Remote-User`, `Remote-Name`, `Remote-Groups`, подделать их снаружи нельзя
- [ ] Как завести учётку и сменить пароль, описано в `infra/deploy/README.md`

## Тест-кейсы

1. Без входа `https://shugraph.feytox.ru` и `https://pr-<N>.shugraph.feytox.ru` перенаправляют на страницу входа.
2. Под аналитиком открывается интерфейс аналитика, а админка и `/admin-api` отвечают отказом.
3. Под админом открываются и интерфейс аналитика, и админка.
4. Запрос с заголовком `Remote-User` снаружи без входа не проходит.
~~~
