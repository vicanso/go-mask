# go-mask

mask and cut value for struct. 

```go
m := mask.New(
    mask.RegExpOption(regexp.MustCompile("password")),
    mask.MaxLengthOption(3),
)
result := m.Struct(struct {
    Name string `json:"name"`
    Password string `json:"password"`
}{
    "test",
    "password",
})
```
