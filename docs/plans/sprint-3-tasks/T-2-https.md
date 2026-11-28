Summary: infra: HTTPS на prod и stage-площадках

- Type: Task
- Priority: Major
- Assignee: feytox
- Tags: infra
- Parent: T-1
- Soft Deadline: 04.12

~~~
## Описание

prod и stage-площадки открываются по HTTPS с сертификатом Let's Encrypt, а HTTP перенаправляется на HTTPS. Без HTTPS пароль при входе пойдёт открытым текстом, а Brave в режиме HTTPS-only сайт сейчас не открывает совсем.

## Решение

- [ ] cert-manager в k3s выпускает и продлевает wildcard-сертификат на `shugraph.feytox.ru` и `*.shugraph.feytox.ru` через DNS-01 в Cloudflare, как в плане из `docs/rfc/01-zeroth-rfc/ci.md`
- [ ] Порт 443 открыт в группе безопасности ВМ, Gateway отдаёт на нём сертификат для prod и всех `pr-<N>`
- [ ] Запросы по HTTP получают редирект на HTTPS
- [ ] Токен Cloudflare лежит в k8s Secret, как его выдать и положить, описано в `infra/deploy/README.md`
- [ ] Ссылки на площадки в `infra/deploy/README.md` и доках ведут на `https://`

## Тест-кейсы

1. `https://shugraph.feytox.ru` и `https://pr-<N>.shugraph.feytox.ru` открываются в Brave в режиме HTTPS-only без предупреждений.
2. `curl -I http://shugraph.feytox.ru` отвечает редиректом на `https://`.
3. Новая stage-площадка открывается по HTTPS сразу после деплоя, без отдельного выпуска сертификата.
~~~
