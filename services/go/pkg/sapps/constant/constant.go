package constant

import (
	"math/rand"
	"os"
	"strings"
)

var (
	Test         = os.Getenv("TEST") == "true"
	API_URL      = "https://ilhan-2.aiplaylist.co"
	KIA_API_KEYS = strings.Split(os.Getenv("KIA_API_KEYS"), ",")
	WD_PATH      = os.Getenv("WD_PATH")
)

func GetKiaAPIKey() string {
	return KIA_API_KEYS[rand.Intn(len(KIA_API_KEYS))]
}
