package i18n

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type LocaliserImpl struct {
	translations map[string]map[string]string
}

func NewLocaliser() (Localiser, error) {
	l := &LocaliserImpl{
		translations: make(map[string]map[string]string),
	}

	files, err := filepath.Glob("i18n/locales/*.json")
	if err != nil {
		return nil, fmt.Errorf("error finding language files: %v", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no language files found")
	}

	for _, file := range files {
		lang := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error reading language file %s: %v", file, err)
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return nil, fmt.Errorf("error parsing language file %s: %v", file, err)
		}

		l.translations[lang] = translations
	}

	return l, nil
}

func (l *LocaliserImpl) GetLocalString(lang, key string, args map[string]string) string {
	langTranslations, ok := l.translations[lang]
	if !ok {
		slog.Error("Language not found",
			slog.String("lang", lang),
		)
		return key
	}

	translation, ok := langTranslations[key]
	if !ok {
		slog.Error("Language key not found",
			slog.String("lang", lang),
			slog.String("key", key),
		)
		return key
	}

	for k, v := range args {
		translation = strings.ReplaceAll(translation, k, v)
	}

	return translation
}
