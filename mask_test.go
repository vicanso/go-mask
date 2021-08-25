package mask

import (
	"encoding/json"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

type data struct {
	Name      string     `json:"name"`
	Count     int        `json:"count"`
	ArrPoint  []*subData `json:"arrPoint"`
	Arr       []subData  `json:"arr"`
	Data      subData    `json:"data"`
	DataPoint *subData   `json:"dataPoint"`
}

type subData struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Max     int    `json:"max"`
}

func TestMaskStruct(t *testing.T) {
	assert := assert.New(t)

	d := data{
		Name:  "test",
		Count: 1,
		ArrPoint: []*subData{
			{
				Title:   "1",
				Content: "Go即将支持泛型",
				Max:     1,
			},
			{
				Title:   "2",
				Content: "测试",
				Max:     2,
			},
		},
		Arr: []subData{
			{
				Title:   "3",
				Content: "test",
				Max:     3,
			},
			{
				Title: "4",
				Max:   4,
			},
		},
		Data: subData{
			Title: "5",
			Max:   5,
		},
		DataPoint: &subData{
			Title: "6",
			Max:   6,
		},
	}

	m := New(
		RegExpOption(regexp.MustCompile("title")),
		MaxLengthOption(4),
	)

	result, err := m.Struct(d)
	assert.Nil(err)
	assert.NotNil(result)
	buf, _ := json.Marshal(result)
	assert.Equal(`{"arr":{"0":{"content":"test","max":3,"title":"***"},"1":{"content":"","max":4,"title":"***"}},"arrPoint":{"0":{"content":"Go即将 ... (4 more runes)","max":1,"title":"***"},"1":{"content":"测试","max":2,"title":"***"}},"count":1,"data":{"content":"","max":5,"title":"***"},"dataPoint":{"content":"","max":6,"title":"***"},"name":"test"}`, string(buf))
}

func TestMaskURLValues(t *testing.T) {
	assert := assert.New(t)

	m := New(
		RegExpOption(regexp.MustCompile("title")),
		MaxLengthOption(4),
	)

	data := url.Values{
		"title": {
			"1",
			"2",
		},
		"category": {
			"测试人员分类",
			"cat",
		},
	}
	result := m.URLValues(data)
	buf, _ := json.Marshal(result)
	assert.Equal(`{"category":["测试人员 ... (2 more runes)","cat"],"title":"***"}`, string(buf))
}

func BenchmarkMaskStruct(b *testing.B) {
	d := data{
		Name:  "test",
		Count: 1,
		ArrPoint: []*subData{
			{
				Title:   "1",
				Content: "Go即将支持泛型",
				Max:     1,
			},
			{
				Title:   "2",
				Content: "测试",
				Max:     2,
			},
		},
		Arr: []subData{
			{
				Title:   "3",
				Content: "test",
				Max:     3,
			},
			{
				Title: "4",
				Max:   4,
			},
		},
		Data: subData{
			Title: "5",
			Max:   5,
		},
		DataPoint: &subData{
			Title: "6",
			Max:   6,
		},
	}
	m := New(
		RegExpOption(regexp.MustCompile("title")),
		MaxLengthOption(4),
	)
	for i := 0; i < b.N; i++ {
		_, err := m.Struct(&d)
		if err != nil {
			panic(err)
		}
	}
}
