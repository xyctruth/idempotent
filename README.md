# Idempotent

Go idempotent library

## GORM Drive

```golang
db := NewDB("root:@tcp(localhost:3306)/test?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4")
i, err := idempotent.New(drive_gorm.New(db))
if err != nil {
    panic(err)
}
_ = db.Transaction(func(tx *gorm.DB) error {
    ok := i.Acquire("idempotent_key", drive_gorm.New(tx), idempotent.WithTTL(time.Minute))
    if !ok {
        fmt.Println("Repeated")
        return nil
    }
    // Some of your code
    fmt.Println("Hello")
    return nil
})
```
