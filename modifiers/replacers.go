package modifiers

import "regexp"

var urlRegex *regexp.Regexp = regexp.MustCompile(
	`(?:https?|jdbc):\/\/(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,6}(?::\d+)?(?:\/[^\s]*)?`,
)

type Replacers interface {
	replace(string) (string, error)
	isEnabled() bool
}

type UrlReplacer struct {
	isEnable     bool
	isApplied    bool
	replacements map[string]string
}

func (urlReplacer *UrlReplacer) replace(string) (string, error) {
	return "", nil
}

func (urlReplacer *UrlReplacer) GetName() string {
	return "URL Replacer"
}

func (urlReplacer *UrlReplacer) isEnabled() bool {
	return urlReplacer.isEnable
}

func (urlModifier *UrlReplacer) Detect(input string) error {
	urls := urlRegex.FindAllString(input, -1)

	for _, url := range urls {
		urlModifier.replacements[url] = url
	}
	return nil
}

func (urlReplacer *UrlReplacer) Modify(input string) (string, error) {
	return "", nil
}

func (urlReplacer *UrlReplacer) GetDetections() map[string]string {
	return urlReplacer.replacements
}

func (urlReplacer *UrlReplacer) IsApplied() bool {
	return urlReplacer.isApplied
}

func NewUrlReplacer() *UrlReplacer {
	return &UrlReplacer{
		isEnable:     true,
		replacements: make(map[string]string),
	}
}
