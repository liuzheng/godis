package godis

import (
    "fmt"
    "time"
    "testing"
    "reflect"
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
    fmt.Println(reflect.DeepEqual(q, &Queue{}))
    fmt.Println(reflect.TypeOf(&Queue{}))
}

type Channel struct {
    value interface{}
    ch    chan interface{}
}

func Test2(t *testing.T) {
    ch1 := make(chan Channel)
    ch2 := make(chan interface{})
    ch3 := make(chan interface{})
    go func() {
        for i := 0; i < 10; i++ {
            ch1 <- Channel{i, ch2}
        }
    }()
    go func() {
        for i := 11; i < 20; i++ {
            ch1 <- Channel{i, ch3}
        }
    }()

    go func() {
        for {
            select {
            case m := <-ch1:
                m.ch <- m.value
            }
        }
    }()
    go func() {
        for {
            select {
            case m := <-ch2:
                fmt.Println("two:", m)
            }
        }
    }()
    go func() {
        for {
            select {
            case m := <-ch3:
                fmt.Println("three:", m)
            }
        }
    }()
        time.Sleep(12 * time.Second)

}