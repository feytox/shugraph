# Граф связей для выявлений мошеннических групп

## Git-хуки

В репозитории настроен [lefthook](https://github.com/evilmartians/lefthook). Хук `commit-msg` добавляет к сообщению коммита префикс задачи из имени ветки: на ветке `GSD-N` коммит `feat: x` превращается в `[GSD-N] feat: x`. На `main` и ветках с другими именами сообщение не меняется.

```bash
# любой из вариантов: https://lefthook.dev/install/
brew install lefthook
go install github.com/evilmartians/lefthook/v2@latest
pnpm setup && pnpm add -g lefthook

lefthook install
```
