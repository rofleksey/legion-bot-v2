package i18n

type Localiser interface {
	GetLocalString(lang, key string, args map[string]string) string
}
