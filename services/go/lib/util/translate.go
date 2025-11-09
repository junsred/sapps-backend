package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/slongfield/pyfmt"
)

var translatedData = map[string]map[string]string{
	"tr": {},
	"en": {},
	"de": {},
	"es": {},
	"fr": {},
	"it": {},
}

func loadLang(lang string) {
	path := fmt.Sprintf("assets/languages/%s.json", lang)
	dat, err := os.ReadFile(path)
	if err != nil {
		if lang == "en" || lang == "tr" {
			log.Panicln(err, path)
		}
		return
	}
	data := make(map[string]string)
	err = json.Unmarshal(dat, &data)
	if err != nil {
		log.Panicln(err)
	}
	if translatedData[lang] == nil {
		translatedData[lang] = make(map[string]string)
	}
	for k, v := range data {
		translatedData[lang][k] = v
	}
}

func LoadFolder() {
	availableLanguages := []string{"en", "tr", "de", "es", "fr", "it"}
	for _, lang := range availableLanguages {
		loadLang(lang)
	}
}

func getKeyValue(lang string, key string) string {
	lang = strings.ToLower(lang)
	langData, ok := translatedData[lang]
	if !ok {
		langData = translatedData["en"]
	}
	strData := langData[key]
	if strData == "" {
		strData = translatedData["en"][key]
		if strData == "" {
			return key
		}
	}
	return strData

}
func GetTranslation(lang string, key string, vals ...interface{}) string {
	strData := getKeyValue(lang, key)
	if len(vals) == 0 {
		return strData
	}
	if len(strData) == 0 {
		return key
	}
	return fmt.Sprintf(strData, vals...)
}

func GetTranslationPythonic(lang string, key string, vals ...interface{}) string {
	strData := getKeyValue(lang, key)
	if len(vals) == 0 {
		return strData
	}
	s, _ := pyfmt.Fmt(strData, vals...)
	return s
}
