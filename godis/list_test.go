package godis

import (
    "fmt"
    "time"
    "testing"
    "reflect"
    "encoding/json"
)

func Test1(t *testing.T) {
    key := "test"
    db := 0
    q := NewQueue(key, db)
    go func() {
        q.ListLpush("one")
    }()
    go func() {
        q.ListLpush("four")
    }()
    q.ListLpush("two")
    q.ListLpush("three")
    v := q.ListLpop()
    fmt.Println("pop v:", v)
    fmt.Println("......")
    time.Sleep(1 * time.Second)
    q.dump()
    fmt.Println(reflect.DeepEqual(q,&Queue{}))
    fmt.Println(reflect.TypeOf(&Queue{}))
    json.Marshal()

}