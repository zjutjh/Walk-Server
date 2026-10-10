package comm

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	playgroundValidator "github.com/go-playground/validator/v10"
	myjwt "github.com/zjutjh/mygo/jwt"
)

var (
	hongKongPermitPattern = regexp.MustCompile(`^H(?:[0-9]{8}|[0-9]{10})$`)
	macaoPermitPattern    = regexp.MustCompile(`^M(?:[0-9]{8}|[0-9]{10})$`)
	taiwanPermitPattern   = regexp.MustCompile(`^(?:[0-9]{8}|[0-9]{10}(?:[A-Z0-9]{1,2}|\([A-Z]\)|\([0-9]{2}\))?)$`)
	chinaPassportPattern  = regexp.MustCompile(`^(?:E[0-9]{8}|E[A-HJ-NP-Z][0-9]{7}|G[0-9]{8}|(?:PE|SE|DE)[0-9]{7}|[PSD][0-9]{8})$`)
	passportSerialPattern = regexp.MustCompile(`^[A-Z0-9]+$`)
	oldPermanentIDPattern = regexp.MustCompile(`^[A-Z]{3}[0-9]{12}$`)
	identityValidator     = playgroundValidator.New()
)

var mainlandProvinceCodes = map[string]struct{}{
	"11": {}, "12": {}, "13": {}, "14": {}, "15": {},
	"21": {}, "22": {}, "23": {},
	"31": {}, "32": {}, "33": {}, "34": {}, "35": {}, "36": {}, "37": {},
	"41": {}, "42": {}, "43": {}, "44": {}, "45": {}, "46": {},
	"50": {}, "51": {}, "52": {}, "53": {}, "54": {},
	"61": {}, "62": {}, "63": {}, "64": {}, "65": {},
	"71": {}, "81": {}, "82": {},
}

var compoundSurnames = map[string]struct{}{
	"欧阳": {}, "司马": {}, "上官": {}, "诸葛": {}, "东方": {},
	"皇甫": {}, "尉迟": {}, "公孙": {}, "慕容": {}, "宇文": {},
	"长孙": {}, "司徒": {}, "司空": {},
}

func GenerateToken(userID int64) (string, error) {
	return myjwt.Pick[string]().GenerateToken(strconv.FormatInt(userID, 10))
}

func GetUserIDFromCtx(ctx *gin.Context) (int64, error) {
	id, err := myjwt.GetIdentity[string](ctx)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(id, 10, 64)
}

func NormalizePhone(phone string) string {
	return strings.TrimSpace(phone)
}

// MaskPersonName 将姓名转换为适合公开展示的脱敏名称。
func MaskPersonName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "队长"
	}

	if isLatinName(name) {
		words := strings.Fields(name)
		masked := make([]string, 0, len(words))
		for _, word := range words {
			runes := []rune(word)
			if len(runes) > 0 {
				masked = append(masked, string(runes[0])+"**")
			}
		}
		return strings.Join(masked, " ")
	}

	runes := []rune(name)
	if len(runes) == 1 {
		return "某同学"
	}
	if len(runes) >= 2 {
		surname := string(runes[:2])
		if _, ok := compoundSurnames[surname]; ok {
			return surname + "同学"
		}
	}
	return string(runes[0]) + "同学"
}

func isLatinName(name string) bool {
	hasLetter := false
	for _, r := range name {
		switch {
		case unicode.IsSpace(r):
			continue
		case unicode.Is(unicode.Latin, r):
			hasLetter = true
		default:
			return false
		}
	}
	return hasLetter
}

func IsValidPhone(phone string) bool {
	phone = NormalizePhone(phone)
	return len(phone) == 11 && allASCIIDigits(phone)
}

func IsValidIdentity(identity string) bool {
	return IsValidIdentityForHome(identity, HomeMainland)
}

// IsValidIdentityForHome 根据户籍类型校验证件号码。
// home 只决定校验规则，不需要持久化到数据库。
func IsValidIdentityForHome(identity string, home HomeType) bool {
	identity = NormalizeIdentity(identity)
	switch home {
	case HomeMainland:
		return isValidMainlandIdentity(identity)
	case HomeHongKongMacao:
		return hongKongPermitPattern.MatchString(identity) || macaoPermitPattern.MatchString(identity)
	case HomeTaiwan:
		return taiwanPermitPattern.MatchString(identity)
	case HomeInternational:
		return isValidPassport(identity) || isValidForeignPermanentIdentity(identity)
	default:
		return false
	}
}

func isValidMainlandIdentity(identity string) bool {
	if len(identity) != 18 || !allASCIIDigits(identity[:17]) {
		return false
	}
	if !isValidMainlandProvinceCode(identity[:2]) || identity[:6] == "000000" {
		return false
	}
	if !isValidIdentityBirthDate(identity[6:14]) || identity[14:17] == "000" {
		return false
	}
	return hasValidMOD112Checksum(identity)
}

func isValidPassport(identity string) bool {
	if chinaPassportPattern.MatchString(identity) {
		return true
	}
	if len(identity) < 4 || len(identity) > 12 || !passportSerialPattern.MatchString(identity[3:]) {
		return false
	}
	return identityValidator.Var(identity[:3], "iso3166_1_alpha3") == nil
}

func isValidForeignPermanentIdentity(identity string) bool {
	if oldPermanentIDPattern.MatchString(identity) {
		return identityValidator.Var(identity[:3], "iso3166_1_alpha3") == nil
	}
	if len(identity) != 18 || identity[0] != '9' || !allASCIIDigits(identity[:17]) {
		return false
	}
	if !isValidMainlandProvinceCode(identity[1:3]) || !isValidIdentityBirthDate(identity[6:14]) {
		return false
	}
	if identity[14:17] == "000" {
		return false
	}
	return hasValidMOD112Checksum(identity)
}

func isValidIdentityBirthDate(value string) bool {
	if len(value) != 8 || value < "19010101" || value > "20991231" {
		return false
	}
	_, err := time.Parse("20060102", value)
	return err == nil
}

func isValidMainlandProvinceCode(value string) bool {
	_, ok := mainlandProvinceCodes[value]
	return ok
}

func hasValidMOD112Checksum(identity string) bool {
	weights := [...]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checksumCodes := "10X98765432"
	sum := 0
	for index, weight := range weights {
		sum += int(identity[index]-'0') * weight
	}
	return identity[17] == checksumCodes[sum%11]
}

func allASCIIDigits(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}
