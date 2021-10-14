# go-mask

mask and cut value for struct. 

```go
m := mask.New(
    mask.RegExpOption(regexp.MustCompile("password")),
    mask.MaxLengthOption(3),
    CustomMaskOption(regexp.MustCompile("desc"), , func(key, value string) string {
        max := 10
        if len(value) <= max {
            return value
        }
        return value[0:max] + "..."
    }),
)
result := m.Struct(struct {
    Name string `json:"name"`
    Password string `json:"password"`
    Desc string `json:"desc"`
}{
    "test",
    "password",
    "desc",
})
```
