package mask

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/tidwall/gjson"
)

type CustomMask struct {
	Reg     *regexp.Regexp
	Handler MaskHandler
}
type Mask struct {
	NotMaskReg  *regexp.Regexp
	Reg         *regexp.Regexp
	MaxLength   int
	CustomMasks []*CustomMask
}

const maskStar = "***"

type MaskOption func(*Mask)
type MaskHandler func(key, value string) string

func New(opts ...MaskOption) *Mask {
	m := &Mask{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// RegExpOption sets regexp option
func RegExpOption(reg *regexp.Regexp) MaskOption {
	return func(m *Mask) {
		m.Reg = reg
	}
}

// NotMaskRegExpOption sets not mask regexp option
func NotMaskRegExpOption(reg *regexp.Regexp) MaskOption {
	return func(m *Mask) {
		m.NotMaskReg = reg
	}
}

// CustomMaskOption add custom mask regexp option
func CustomMaskOption(reg *regexp.Regexp, handler MaskHandler) MaskOption {
	return func(m *Mask) {
		if len(m.CustomMasks) == 0 {
			m.CustomMasks = make([]*CustomMask, 0)
		}
		m.CustomMasks = append(m.CustomMasks, &CustomMask{
			Reg:     reg,
			Handler: handler,
		})
	}
}

// Set max string length option for mask, if the length of rune is gt max, it will convert to %s ... (%d more runes)
// 0 means no limit.
func MaxLengthOption(maxLength int) MaskOption {
	return func(m *Mask) {
		m.MaxLength = maxLength
	}
}

func (m *Mask) cutString(str string) string {
	if m.MaxLength <= 0 {
		return str
	}
	// 如果超5倍长度，直接截短
	if len(str) > 5*m.MaxLength {
		return fmt.Sprintf("%s ... (%d more strings)", str[0:m.MaxLength], len(str)-m.MaxLength)
	}
	r := []rune(str)
	moreRunes := len(r) - m.MaxLength
	if moreRunes > 0 {
		return fmt.Sprintf("%s ... (%d more runes)", string(r[0:m.MaxLength]), moreRunes)
	}
	return str
}

func (m *Mask) convert(result *gjson.Result) interface{} {
	data := make(map[string]interface{})
	arr := make([]interface{}, 0)
	isArray := result.IsArray()
	appendLog := func(key string, value interface{}) {
		if isArray {
			arr = append(arr, value)
		} else {
			data[key] = value
		}
	}
	result.ForEach(func(key, value gjson.Result) bool {
		k := key.String()
		if m.NotMaskReg != nil && m.NotMaskReg.MatchString(k) {
			appendLog(k, value.Value())
			return true
		}
		// 如果能匹配则使用 ***
		if m.Reg != nil && m.Reg.MatchString(k) {
			appendLog(k, maskStar)
			return true
		}
		for _, customMask := range m.CustomMasks {
			// 如果符合自定义的mask处理
			if customMask.Reg.MatchString(k) {
				appendLog(k, customMask.Handler(k, value.String()))
				return true
			}
		}
		if value.IsObject() || value.IsArray() {
			appendLog(k, m.convert(&value))
			return true
		}
		// 如果限制最大长度
		if m.MaxLength != 0 && value.Type == gjson.String {
			appendLog(k, m.cutString(value.String()))
		} else {
			appendLog(k, value.Value())
		}
		return true
	})
	if isArray {
		return arr
	}
	return data
}

// Struct converts struct to map[string]interface{} with mask
func (m *Mask) Struct(data interface{}) (map[string]interface{}, error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	result := gjson.ParseBytes(buf)
	if result.IsArray() {
		return nil, errors.New("array is not supported")
	}
	tmp, _ := m.convert(&result).(map[string]interface{})
	return tmp, nil
}

// URLValues converts url values to map[string]interface{} with mask
func (m *Mask) URLValues(data url.Values) map[string]interface{} {
	result := make(map[string]interface{})
	for key, values := range data {
		if m.NotMaskReg != nil && m.NotMaskReg.MatchString(key) {
			result[key] = values
			continue
		}
		if m.Reg != nil && m.Reg.MatchString(key) {
			result[key] = maskStar
			continue
		}
		matchCustomMask := false
		for _, customMask := range m.CustomMasks {
			if customMask.Reg.MatchString(key) {
				arr := make([]string, len(values))
				for index, value := range values {
					arr[index] = customMask.Handler(key, value)
				}
				result[key] = arr
				matchCustomMask = true
				break
			}
		}
		// 如果已经匹配自定义的mask处理
		if matchCustomMask {
			continue
		}

		arr := make([]string, len(values))
		for index, value := range values {
			arr[index] = m.cutString(value)
		}
		result[key] = arr
	}
	return result
}
