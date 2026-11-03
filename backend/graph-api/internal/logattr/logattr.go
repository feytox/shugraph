package logattr

import "log/slog"

func Error(err error) slog.Attr {
	return slog.Any("error.message", err)
}

func Stack(stack []byte) slog.Attr {
	return slog.String("error.stack", string(stack))
}
