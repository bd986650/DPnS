package ctxlang

import (
	"context"
	"strings"
)

type langKey struct{}

// ParseAcceptLanguage разбирает заголовок Accept-Language (первый тег).
func ParseAcceptLanguage(h string) string {
	parts := strings.Split(h, ",")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "en"
	}
	tag := strings.TrimSpace(strings.Split(parts[0], ";")[0])
	tag = strings.ToLower(tag)
	if strings.HasPrefix(tag, "ru") {
		return "ru"
	}
	return "en"
}

// With сохраняет код языка ответа (en | ru) в контексте.
func With(ctx context.Context, lang string) context.Context {
	if lang != "ru" {
		lang = "en"
	}
	return context.WithValue(ctx, langKey{}, lang)
}

// From возвращает язык из контекста или "en".
func From(ctx context.Context) string {
	if v, ok := ctx.Value(langKey{}).(string); ok && v == "ru" {
		return "ru"
	}
	return "en"
}
