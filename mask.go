package mask

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/tidwall/gjson"
)

type Mask struct {
	Reg       *regexp.Regexp
	MaxLength int
}

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
			data[k] = "***"
		} else if value.IsObject() || value.IsArray() {
			data[k] = m.convert(&value)
		} else {
			// 如果限制最大长度
			if m.MaxLength != 0 && value.Type == gjson.String {
				str := value.String()
				r := []rune(value.String())
				moreRunes := len(r) - m.MaxLength
				if moreRunes > 0 {
					str = fmt.Sprintf("%s ... (%d more runes)", string(r[0:m.MaxLength]), moreRunes)
				}
				data[k] = str
			} else {
				data[k] = value.Value()
			}
		}
		index++
		return true
	})
	return data
}

// Convert struct to map[string]interface{} with mask
func (m *Mask) Struct(data interface{}) (map[string]interface{}, error) {
	buf, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	result := gjson.ParseBytes(buf)
	return m.convert(&result), nil
}