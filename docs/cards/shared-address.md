# Карточка расследования: «Один адрес на толпу»

Шаблон текста карточки для [Ф5](../features.md#ф5-карточка-расследования), схема «адрес» (`scheme: address`).

Фронт собирает объяснение из `risk_parts.address` и `features` группы, которые приходят в ответе graph-api.

## Данные для сборки

| Что | Откуда в ответе API | Пример |
|---|---|---|
| Уровень риска | `risk_level` | `medium` |
| Итоговый риск | `risk_score` | `65` |
| Слагаемые риска | `risk_parts.address` | `{"scheme": 40, "speed": 10, "clients": 0, "devices": 5, "burst": 10}` |
| Длительность | `features.span_minutes` | `40` |
| Число счетов | `features.accounts_count` | `5` |
| Оборот | `features.total_amount`, копейки | `19500000` |
| Число клиентов | `logins` группы, уникальные `client_id` | `4` |
| Число устройств | `logins` группы, уникальные `device_id` | `3` |
| Кучность первых входов | `logins` группы, разница между максимальным и минимальным первым входом клиентов (в минутах) | `10` |

`risk_parts.address` содержит баллы, а не исходные количества: `clients` — баллы за число клиентов, `devices` — баллы за число устройств, `burst` — баллы за кучность первых входов.

`clients_count` — число различных `client_id` в `logins` группы.
`devices_count` — число различных `device_id` в `logins` группы.
`burst_minutes` — время в целых минутах между самым ранним и самым поздним первым входом клиента в окне группы.

Подписи уровня риска:

| `risk_level` | Подпись |
|---|---|
| `low` | низкий |
| `medium` | средний |
| `high` | высокий |

## Слагаемые риска

| Слагаемое из [risk.md](../schemes/risk.md) | Плейсхолдер в шаблоне | Что подставляем | Формулировка |
|---|---|---|---|
| Тип схемы «адрес» | `scheme_line` | `risk_parts.address.scheme` | `Тип схемы «адрес»: +{scheme}.` |
| Скорость | `speed_line` | `features.span_minutes`, `risk_parts.address.speed` | `Скорость: {span_minutes} мин от первого до последнего события — +{speed}.` |
| Число клиентов | `clients_line` | `clients_count`, `risk_parts.address.clients` | `Клиентов на одном адресе: {clients_count} — +{clients}.` |
| Число устройств | `devices_line` | `devices_count`, `risk_parts.address.devices` | `Устройств на одном адресе: {devices_count} — +{devices}.` |
| Кучность первых входов | `burst_line` | `burst_minutes`, `risk_parts.address.burst` | `Кучность первых входов: {burst_minutes} мин — +{burst}.` |

Строки обоснования собираются по таблице; шаблон только подставляет соответствующие плейсхолдеры. `risk_parts.address.clients`, `devices` и `burst` — баллы по [risk.md](../schemes/risk.md); в тексте подставляем фактические `clients_count`, `devices_count` и `burst_minutes`.

Итог берётся из `risk_score`; для схемы сумма слагаемых не превышает 100, поэтому совпадает с итогом.

## Шаблон текста

```text
Риск: {risk_level_label} ({risk_score}/100)
Счетов в группе: {accounts_count}, оборот: {total_amount_rub} ₽.

Что произошло:
{scheme_summary}

Обоснование риска:
- {scheme_line}
- {speed_line}
- {clients_line}
- {devices_line}
- {burst_line}

Итого: {risk_score}/100 — {risk_level_label}.
```

`total_amount_rub` — `features.total_amount / 100`, целые рубли с разделителем разрядов.

`scheme_summary` — короткое объяснение схемы простыми словами:

`С одного сетевого адреса заходили в личный кабинет разные клиенты (всего {clients_count}) с разных устройств (всего {devices_count}), причём первые входы уложились в {burst_minutes} мин. Деньги с их счетов переведены на счета вывода. Это признак схемы с общим адресом: разные клиенты и устройства выходят в сеть из одной точки, будто за операциями стоит один координатор.`


## Сценарий `shared_address`

Сценарий: [`backend/graph-admin-api/scenarios/shared_address.json`](../../backend/graph-admin-api/scenarios/shared_address.json). После запуска группа должна прийти с такими полями:

| Поле | Значение |
|---|---|
| `scheme` | `address` |
| `risk_score` | `65` |
| `risk_level` | `medium` |
| `risk_parts.address` | `{"scheme": 40, "speed": 10, "clients": 0, "devices": 5, "burst": 10}` |
| `features.span_minutes` | `40` |
| `features.accounts_count` | `5` |
| `features.total_amount` | `19500000` |
| `clients_count` | `4` |
| `devices_count` | `3` |
| `burst_minutes` | `10` |

Проверка чисел: `40 + 10 + 0 + 5 + 10 = 65`, уровень `medium` — диапазон 60–74. `span_minutes` — 40 минут (18:00–18:40), `total_amount` — 19 500 000 копеек (50 000 + 45 000 + 40 000 + 60 000 ₽). `accounts_count` — 5 счетов, `clients_count` — 4 различных `client_id` в `logins`, `devices_count` — 3 различных `device_id` в `logins`, `burst_minutes` — 10 минут (18:00–18:10).

Ожидаемый текст карточки:

```text
Риск: средний (65/100)
Счетов в группе: 5, оборот: 195 000 ₽.

Что произошло:
С одного сетевого адреса заходили в личный кабинет разные клиенты (всего 4) с разных устройств (всего 3), причём первые входы уложились в 10 мин. Деньги с их счетов переведены на счета вывода. Это признак схемы с общим адресом: разные клиенты и устройства выходят в сеть из одной точки, будто за операциями стоит один координатор.

Обоснование риска:
- Тип схемы «адрес»: +40.
- Скорость: 40 мин от первого до последнего события — +10.
- Клиентов на одном адресе: 4 — +0.
- Устройств на одном адресе: 3 — +5.
- Кучность первых входов: 10 мин — +10.

Итого: 65/100 — средний.
```
