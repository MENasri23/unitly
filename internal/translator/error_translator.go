package translator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"unitly/internal/domain"
)

type ErrorTranslator struct {
	defaultLanguage string
	messages        map[string]map[domain.ErrorCode]string
}

func NewErrorTranslator(localeFS fs.FS, defaultLang string) Translator {
	return &ErrorTranslator{
		defaultLanguage: defaultLang,
		messages:        make(map[string]map[domain.ErrorCode]string),
	}
}

// loadJSONLocales walks the given fs.FS directory and unmarshals JSON files into domainMessages
func (t *ErrorTranslator) loadJSONLocales(localeFS fs.FS) error {
	return fs.WalkDir(localeFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-JSON files
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		// Extract language tag from file name (e.g., "locales/en.json" -> "en")
		filename := filepath.Base(path)
		lang := strings.TrimSuffix(filename, filepath.Ext(filename))

		data, err := fs.ReadFile(localeFS, path)
		if err != nil {
			return fmt.Errorf("unable to read file %s: %w", path, err)
		}

		var messages map[domain.ErrorCode]string
		if err := json.Unmarshal(data, &messages); err != nil {
			return fmt.Errorf("unable to parse JSON file %s: %w", path, err)
		}

		t.messages[lang] = messages
		return nil
	})
}

func (t *ErrorTranslator) Translate(lang string, err error) string {
	var domainErr domain.Error
	if !errors.As(err, &domainErr) {
		return err.Error()
	}

	langMap, exists := t.messages[lang]
	if !exists {
		langMap = t.messages[t.defaultLanguage]
	}

	msg, found := langMap[domainErr.Code()]
	if !found {
		return err.Error()
	}

	// Dynamic parameters interpolation: e.g. "{entity} not found" -> "user not found"
	if domainErr.Params() != nil {
		for k, v := range domainErr.Params() {
			placeholder := fmt.Sprintf("{%s}", k)
			msg = strings.ReplaceAll(msg, placeholder, fmt.Sprintf("%v", v))
		}
	}

	return msg
}
