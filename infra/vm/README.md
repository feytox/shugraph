# ВМ с k3s

Одна прерываемая ВМ в Яндекс Облаке, на ней single-node k3s.

| Параметр     | Значение                                          |
|--------------|---------------------------------------------------|
| Имя          | `shucluster-vm`                                   |
| Зона         | ru-central1-d                                     |
| Платформа    | standard-v3, прерываемая                          |
| vCPU         | 2, гарантированная доля 20%                       |
| RAM          | 4 ГБ                                              |
| Диск         | HDD 20 ГБ, Ubuntu 24.04 LTS                       |
| Публичный IP | динамический, меняется при каждом запуске         |
| Порты        | входящие 22 и 80, исходящие все                   |
| SSH          | пользователь `shugraph`, ключ `shugraph-ci`       |
| k3s          | v1.36.5+k3s1                                      |
| kubectl      | `ssh shugraph@<ip> sudo k3s kubectl`, 6443 закрыт |

- Спеки упираются в квоты облака: на зону 2 vCPU, 20 ГБ HDD, 10 ГБ SSD. Больше только через поддержку.
- Прерываемая ВМ выключается сама минимум раз в сутки. Данные на диске сохраняются, k3s поднимается сам при запуске.
- Включение и выключение: workflow `up` и `down` или локально `./vm.sh up|down` (если есть ключ).

## Доступ для разработчиков

Все заходят под `shugraph` (sudo без пароля) со своим ключом.

(P.s. сделано для упрощения разработки MVP)

1. Разработчик создаёт отдельный ключ и присылает `~/.ssh/shugraph.pub`:
   ```bash
   ssh-keygen -t ed25519 -C <ник> -f ~/.ssh/shugraph
   ```
2. Ответственный вписывает ключ в `authorized_keys` на ВМ:
   ```bash
   ssh -i ~/.ssh/shugraph-ci shugraph@<ip> 'cat >> ~/.ssh/authorized_keys' < shugraph.pub
   ```
3. Вход: `ssh -i ~/.ssh/shugraph shugraph@<ip>`, IP в summary последнего запуска `up`.

После пересоздания ВМ ключи разработчиков нужно добавить заново, в `cloud-init.yaml` только ключ CI.

## Поднять с нуля

Раздел нужен, только чтобы пересоздать ВМ проекта. Шаг 5 заменяет секреты, с которыми работают `up` и `down`, поэтому, если проверяешь инструкцию в своём облаке, делай это в своём форке, а у `gh secret set` укажи `-R <ник>/shugraph`.

Нужны `yc` (профиль с правами на каталог), `jq` и `gh`. Команды запускаются из `infra/vm/`.

В каталоге должна быть сеть `default` с подсетью `default-ru-central1-d`. Если подсети нет:
```bash
yc vpc subnet create --name default-ru-central1-d --zone ru-central1-d --network-name default --range 10.130.0.0/24
```

1. Создать SSH-ключ и вставить публичную часть в `cloud-init.yaml` вместо текущей:
   ```bash
   ssh-keygen -t ed25519 -N "" -C shugraph-ci -f ~/.ssh/shugraph-ci
   ```
2. Создать группу безопасности и ВМ. ВМ сразу запускается и начинает тратить деньги, k3s поставит cloud-init при первом запуске:
   ```bash
   ./create-vm.sh
   ```
3. Вписать ID новой ВМ в `VM_ID` в `vm.sh`:
   ```bash
   yc compute instance get --name shucluster-vm --format json | jq -r .id
   ```
4. Создать сервисный аккаунт для CI с правом только запускать и останавливать эту ВМ:
   ```bash
   yc iam service-account create --name shugraph-ci
   yc compute instance add-access-binding --name shucluster-vm --role compute.operator --service-account-name shugraph-ci
   yc iam key create --service-account-name shugraph-ci --output key.json > /dev/null
   ```
5. Положить ключи в секреты окружения `vm` и удалить `key.json`. Это может сделать только владелец репозитория, секреты обычным коллабораторам недоступны:
   ```bash
   gh api -X PUT 'repos/{owner}/{repo}/environments/vm' > /dev/null
   gh secret set YC_SA_KEY --env vm < key.json && rm key.json
   gh secret set SSH_PRIVATE_KEY --env vm < ~/.ssh/shugraph-ci
   ```
6. Проверить: `./vm.sh up` должен вывести ноду в статусе `Ready` и IP. После проверки выключить ВМ: `./vm.sh down`.

## Удалить

Если `create-vm.sh` упал на середине или ВМ нужно пересоздать, удалить созданное (диск удаляется вместе с ВМ, сервисный аккаунт есть, только если дошли до шага 4):
```bash
yc compute instance delete shucluster-vm
yc vpc security-group delete shucluster-group
yc iam service-account delete shugraph-ci
```

## Стоимость

По калькулятору YC на октябрь 2026, если ВМ работает круглосуточно:

| Ресурс         | Руб./мес |
|----------------|---------:|
| vCPU (2 x 20%) |      239 |
| RAM 4 ГБ       |      239 |
| Публичный IP   |      190 |
| HDD 20 ГБ      |       69 |
| Итого          |      737 |

Включённая ВМ стоит около 0,93 руб./ч. Выключенная стоит около 69 руб./мес (только диск).
