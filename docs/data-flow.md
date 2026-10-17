# Поток данных

Путь одного события: от сценария в админке до экрана аналитика и обратно, после решения аналитика.

## Общая картина

```mermaid
flowchart TB
    subgraph ADM["graph-admin-api"]
        SC[/"сценарий"/] --> PL["проигрыватель:<br/>шлёт события по одному"]
    end

    subgraph CORE["graph-api"]
        IN{"приём"}
        SAVE["записать событие в БД"]
        RB{"пересборка групп:<br/>набрался минимум<br/>по правилам SRD?"}
        PLAIN["событие обычное"]
        GRP["группа создана или дополнена,<br/>событие стало подозрительным"]
        API["API"]
    end

    DB[("PostgreSQL")]

    subgraph FR["graph-frontapp"]
        LIST["список групп"] --> CARD["граф группы"] --> BTN["подтвердить /<br/>отклонить"]
    end

    PL -- "событие" --> IN
    IN -. "битое или дубль:<br/>ничего не пишем" .-> PL
    IN -- "новое" --> SAVE --> RB
    RB -- "нет" --> PLAIN
    RB -- "да" --> GRP
    GRP --> DB
    DB <--> API
    API -- "polling" --> LIST
    BTN -- "решение" --> API

    classDef red fill:#f8d7d7,stroke:#c0392b,color:#222
    classDef blue fill:#d6e6f8,stroke:#2e6db4,color:#222
    class GRP red
    class BTN blue
```

- **Подозрительное событие** - событие, которое попало в группу (`group_events`). Отдельного флага у события нет.
- Фронт получает не все события, а только группы и их подграфы.

## На примере цепочки

Сценарий "быстрая цепочка" из [schemes-simple.md](schemes-simple.md#схема-3-быстрая-цепочка): четыре перевода Олег -> Игорь -> Павел -> Роман -> Тимур.

```mermaid
sequenceDiagram
    participant ADM as graph-admin-api
    participant API as graph-api
    participant DB as PostgreSQL
    participant FR as фронт аналитика

    ADM->>API: evt-0001: Олег → Игорь, 83 500, 18:03
    API->>DB: записать событие
    Note over API: 1 перевод, группы нет
    ADM->>API: evt-0002: Игорь → Павел, 82 000, 18:07
    API->>DB: записать событие
    Note over API: 2 перевода, группы нет

    rect rgba(220, 60, 60, 0.15)
    ADM->>API: evt-0003: Павел → Роман, 80 000, 18:11
    API->>DB: записать событие
    Note over API: 3 перевода за 8 мин, дошло 95%,<br/>сработало правило цепочки (SRD-5)
    API->>DB: новая группа "цепочка", риск 80<br/>события группы: evt-0001...0003
    end

    FR->>API: polling: есть новые группы?
    API-->>FR: цепочка, риск 80, новая

    ADM->>API: evt-0004: Роман → Тимур, 79 000, 18:14
    API->>DB: записать событие
    Note over API: группа дополнена,<br/>4 перевода за 11 мин, дошло 94%
    API->>DB: + evt-0004, риск 85

    ADM->>API: evt-0004 ещё раз (повторный запуск)
    API-->>ADM: дубль, ничего не меняем

    FR->>API: решение аналитика по группе
    API->>DB: новый статус группы
```

Риск 80 = 55 (цепочка) + 20 (8 мин) + 0 (3 перевода) + 5 (дошло 95%). После четвёртого перевода: 55 + 15 + 10 + 5 = 85.

## Решение аналитика

Что graph-api меняет в `groups` после нажатия кнопки:

```mermaid
flowchart LR
    G["группа<br/><b>новая</b>"]
    G -- "подтвердить<br/>+ комментарий" --> C["<b>подтверждена</b><br/>is_frozen = true<br/>decided_at<br/>decision_comment"]
    G -- "отклонить<br/>+ комментарий" --> R["<b>отклонена</b><br/>is_frozen = true<br/>decided_at<br/>decision_comment"]

    classDef red fill:#f8d7d7,stroke:#c0392b,color:#222
    classDef green fill:#d5f0dc,stroke:#27ae60,color:#222
    class C red
    class R green
```

После решения ничего не удаляется: группа остаётся в БД, просто меняет статус. Жизненный цикл группы описан в [schemes-simple.md](schemes-simple.md#жизнь-группы).
