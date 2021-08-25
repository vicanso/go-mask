package mask

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"github.com/tidwall/gjson"
)

type Mask struct {
	Reg       *regexp.Regexp
	MaxLength int
}

const maskStar = "***"

type MaskOption func(*Mask)

func New(opts ...MaskOption) *Mask {
	m := &Mask{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Sets reg exp option for mask
func RegExpOption(reg *regexp.Regexp) MaskOption {
	return func(m *Mask) {
		m.Reg = reg
	}
}

// Set max string length option for mask, if the length of rune is gt max, it will convert to ...
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
	r := []rune(str)
	moreRunes := len(r) - m.MaxLength
	if moreRunes > 0 {
		return fmt.Sprintf("%s ... (%d more runes)", string(r[0:m.MaxLength]), moreRunes)
	}
	return str
}

func (m *Mask) convert(result *gjson.Result) map[string]interface{} {
	data := make(map[string]interface{})
	isArray := result.IsArray()
	index := 0
	result.ForEach(func(key, value gjson.Result) bool {
		k := key.String()
		// 如果是数组则k转换为index
		if isArray {
			k = strconv.Itoa(index)
		}
		// 如果能匹配则使用 ***
		if m.Reg != nil && m.Reg.MatchString(k) {
			data[k] = maskStar
		} else if value.IsObject() || value.IsArray() {
			data[k] = m.convert(&value)
		} else {
			// 如果限制最大长度
			if m.MaxLength != 0 && value.Type == gjson.String {
				data[k] = m.cutString(value.String())
			} else {
				data[k] = value.Value()
			}
		}
		index++
		return true
	})
	return data
}

// Struct converts struct to map[string]interface{} with mask
func (m *Mask) Struct(data interface{}) (map[string]interface{}, error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	result := gjson.ParseBytes(buf)
	return m.convert(&result), nil
}

// URLValues converts url values to map[string]interface{} with mask
func (m *Mask) URLValues(data url.Values) map[string]interface{} {
	result := make(map[string]interface{})
	for key, values := range data {
		if m.Reg != nil && m.Reg.MatchString(key) {
			result[key] = maskStar
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
