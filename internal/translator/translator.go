package translator

type Translator interface {
	Translate(lang string, err error) string
}