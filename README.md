# Граф связей для выявлений мошеннических групп

## Git-хуки

В репозитории настроен [lefthook](https://github.com/evilmartians/lefthook). Хук `prepare-commit-msg` добавляет к сообщению коммита префикс задачи из имени ветки: на ветке `GSD-N` коммит `feat: x` превращается в `[GSD-N] feat: x`. На `main` и ветках с другими именами сообщение не меняется.

```bash
pnpm add -g lefthook && lefthook install
```
