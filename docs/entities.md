# Сущности

Общая табличка:

```mermaid
erDiagram
    CLIENTS {
        string id PK
    }
    ACCOUNTS {
        string id PK
        string client_id FK
    }
    DEVICES {
        string device_id PK
    }
    NETWORK_ADDRESSES {
        string address_id PK
    }
    EVENTS {
        string event_id PK
        enum type
        timestamp occurred_at
    }
    TRANSFERS {
        string event_id PK, FK
        string from_account_id FK
        string to_account_id FK
        money amount
    }
    LOGIN_EVENTS {
        string event_id PK, FK
        string account_id FK
        string client_id FK
        string device_id FK
        string address_id FK
    }
    GROUPS {
        uuid id PK
        enum scheme
        enum status
        integer risk_score
        json risk_parts
        timestamp first_event_at
        timestamp last_event_at
        boolean is_frozen
        string decision_comment
        timestamp? decided_at
    }
    GROUP_MEMBERS {
        uuid group_id PK, FK
        string account_id PK, FK
    }
    GROUP_EVENTS {
        uuid group_id PK, FK
        string event_id PK, FK
    }
    GROUP_FEATURES {
        uuid group_id PK, FK
        integer accounts_count
        integer span_minutes
        money total_amount
    }
    CLIENTS ||--o{ ACCOUNTS : " "
    ACCOUNTS ||--o{ TRANSFERS : " "
    EVENTS ||--|| TRANSFERS : " "
    EVENTS ||--|| LOGIN_EVENTS : " "
    ACCOUNTS ||--o{ LOGIN_EVENTS : " "
    CLIENTS ||--o{ LOGIN_EVENTS : " "
    DEVICES ||--o{ LOGIN_EVENTS : " "
    NETWORK_ADDRESSES ||--o{ LOGIN_EVENTS : " "
    GROUPS ||--o{ GROUP_MEMBERS : " "
    ACCOUNTS ||--o{ GROUP_MEMBERS : " "
    GROUPS ||--o{ GROUP_EVENTS : " "
    EVENTS ||--o{ GROUP_EVENTS : " "
    GROUPS ||--|| GROUP_FEATURES : " "
```

## Клиенты (`clients`)

| Поле | Тип | Откуда берётся | Пример |
|---|---|---|---|
| `id` | string | из события (владелец счёта) | `cli-0421` |

## Счета (`accounts`)

| Поле | Тип | Откуда берётся | Пример |
|---|---|---|---|
| `id` | string | из события (номер счёта отправителя/получателя, счёт входа) | `acc-6742` |
| `client_id` | string | из события (владелец счёта) | `cli-0421` |

## Устройства (`devices`)

| Поле | Тип | Откуда берётся | Пример |
|---|---|---|---|
| `device_id` | string | из события входа | `dev-a91f` |

## Сетевые адреса (`network_addresses`)

| Поле | Тип | Откуда берётся | Пример |
|---|---|---|---|
| `address_id` | string | из события входа (IP) | `84.17.52.10` |

## События (`events`)

| Поле | Тип | Откуда берётся | Пример |
|---|---|---|---|
| `event_id` | string | из события (номер от источника, задаётся в сценарии) | `evt-0001` |
| `type` | enum: перевод / вход | из события | `перевод` |
| `occurred_at` | timestamp | из события (время события) | `2026-10-12T18:03:00Z` |

## Переводы (`transfers`)

| Поле | Тип | Откуда берётся | Пример |
|---|---|---|---|
| `event_id` | string | из события (тот же id) | `evt-0001` |
| `from_account_id` | string | из события (счёт отправителя) | `acc-6742` |
| `to_account_id` | string | из события (счёт получателя) | `acc-6801` |
| `amount` | money | из события (сумма) | `83500.00` |

## События входа (`login_events`)

| Поле | Тип | Откуда берётся | Пример |
|---|---|---|---|
| `event_id` | string | из события | `evt-0002` |
| `account_id` | string | из события (счёт и владелец) | `acc-6742` |
| `client_id` | string | из владельца счёта на момент приёма | `cli-0421` |
| `device_id` | string | из события (отпечаток) | `dev-a91f` |
| `address_id` | string | из события (IP) | `84.17.52.10` |

## Группы (`groups`)

| Поле | Тип | Откуда берётся | Пример |
|---|---|---|---|
| `id` | uuid | системное | `b5f2a1c4-3d6e-4f7a-8b9c-0d1e2f3a4b5c` |
| `scheme` | enum: ферма / адрес / цепочка / кольцо | вычисление | `цепочка` |
| `status` | enum: новая / подтверждена / отклонена | по умолчанию `новая`, дальше действия аналитика| `подтверждена` |
| `risk_score` | integer | вычисление | `85` |
| `risk_parts` | json | вычисление | `{"scheme": 55, "speed": 15, "hops": 10, "preserved": 5}` |
| `first_event_at` | timestamp | вычисляется в приложении на основе событий | `2026-10-12T18:03:00Z` |
| `last_event_at` | timestamp | вычисляется в приложении на основе событий | `2026-10-12T18:14:00Z` |
| `is_frozen` | boolean | решение аналитика или события вышли из окна | `true` |
| `decision_comment` | string | действие аналитика | `транзит за 11 минут, подтверждаю` |
| `decided_at` | timestamp, nullable | действие аналитика (Ф6), NULL пока «новая» | `2026-10-12T18:40:00Z` |

Примеры `risk_parts` по схемам (ключ — слагаемое, сумма значений = `risk_score`):

Ферма:

```json
{"scheme": 50, "speed": 15, "clients": 0, "collector": 10}
```

Адрес:

```json
{"scheme": 40, "speed": 10, "accounts": 0, "devices": 5, "burst": 10}
```

Цепочка:

```json
{"scheme": 55, "speed": 15, "hops": 10, "preserved": 5}
```

Кольцо:

```json
{"scheme": 65, "speed": 10, "rounds": 0, "accounts": 0, "preserved": 10}
```

## Участник группы (`group_members`)

| Поле | Тип | Откуда берётся | Пример |
|---|---|---|---|
| `group_id` | uuid | вычисление | `b5f2a1c4-3d6e-4f7a-8b9c-0d1e2f3a4b5c` |
| `account_id` | string | вычисление | `acc-6742` |

## Доказательства группы (`group_events`)

| Поле | Тип | Откуда берётся | Пример |
|---|---|---|---|
| `group_id` | uuid | вычисление | `b5f2a1c4-3d6e-4f7a-8b9c-0d1e2f3a4b5c` |
| `event_id` | string | вычисление | `evt-0001` |

## Признаки группы (`group_features`)

| Поле | Тип | Откуда берётся | Пример |
|---|---|---|---|
| `group_id` | uuid | вычисление | `b5f2a1c4-3d6e-4f7a-8b9c-0d1e2f3a4b5c` |
| `accounts_count` | integer | вычисление | `5` |
| `span_minutes` | integer | вычисление | `11` |
| `total_amount` | money | вычисление | `324500.00` |

Примечание по `total_amount`: это оборот т.е. сумма всех переводов группы по `group_events` (повторы кругов суммируются)

Схемо-зависимые признаки отдельно не хранятся: разбор риска по слагаемым лежит в `groups.risk_parts`, сырьё для него — события группы (`group_events`)
