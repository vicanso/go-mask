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
	Desc    string `json:"desc"`
	Max     int    `json:"max"`
}

func TestMaskStruct(t *testing.T) {
	assert := assert.New(t)

	d := data{
		Name:  "我的名字测试",
		Count: 1,
		ArrPoint: []*subData{
			{
				Title:   "1",
				Content: "Go即将支持泛型Go即将支持泛型Go即将支持泛型",
				Max:     1,
				Desc:    "string will be cut",
			},
			{
				Title:   "2",
				Content: "测试",
				Max:     2,
				Desc:    "not cut",
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
		CustomMaskOption(regexp.MustCompile("desc"), func(key, value string) string {
			max := 10
			if len(value) <= max {
				return value
			}
			return value[0:max] + "..."
		}),
		RegExpOption(regexp.MustCompile("title")),
		NotMaskRegExpOption(regexp.MustCompile("name")),
		MaxLengthOption(4),
	)

	result, err := m.Struct(d)
	assert.Nil(err)
	assert.NotNil(result)
	buf, _ := json.Marshal(result)
	assert.Equal(`{"arr":[{"content":"test","desc":"","max":3,"title":"***"},{"content":"","desc":"","max":4,"title":"***"}],"arrPoint":[{"content":"Go\ufffd\ufffd ... (56 more strings)","desc":"string wil...","max":1,"title":"***"},{"content":"测试","desc":"not cut","max":2,"title":"***"}],"count":1,"data":{"content":"","desc":"","max":5,"title":"***"},"dataPoint":{"content":"","desc":"","max":6,"title":"***"},"name":"我的名字测试"}`, string(buf))
}

func TestMaskURLValues(t *testing.T) {
	assert := assert.New(t)

	m := New(
		RegExpOption(regexp.MustCompile("title")),
		NotMaskRegExpOption(regexp.MustCompile("name")),
		CustomMaskOption(regexp.MustCompile("category"), func(key, value string) string {
			v := []rune(value)
			return string(v[0])
		}),
		MaxLengthOption(4),
	)

	data := url.Values{
		"name": {
			"我的名字测试",
		},
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
	assert.Equal(`{"category":["测","c"],"name":["我的名字测试"],"title":"***"}`, string(buf))
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
		NotMaskRegExpOption(regexp.MustCompile("name")),
		MaxLengthOption(4),
	)
	for i := 0; i < b.N; i++ {
		_, err := m.Struct(&d)
		if err != nil {
			panic(err)
		}
	}
}
