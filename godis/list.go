package godis

import (
    "container/list"
    "sync"
    "fmt"
)

type Queue struct {
    key    string
    db     int
    data   *list.List
    lock   *sync.Mutex
    T      string
    length int
}

func NewQueue(key string, db int) *Queue {
    return &Queue{
        data:list.New(),
        lock:new(sync.Mutex),
        key:key,
        db:db,
        T:"list",
        length:0,
    }
}

func (q *Queue) ListLpush(v interface{}) int {
    defer q.lock.Unlock()
    q.lock.Lock()
    q.data.PushFront(v)
    q.length++
    return q.length
}

func (q *Queue) ListLpop() interface{} {
    defer q.lock.Unlock()
    q.lock.Lock()
    iter := q.data.Back()
    v := iter.Value
    q.data.Remove(iter)
    return v
}
func (q *Queue) ListLrange(start, stop int) []interface{} {
    defer q.lock.Unlock()
    q.lock.Lock()

    for i := start; i <= stop; i++ {
        fmt.Println(i)
    }
    return nil
}

func (q *Queue) dump() {
    for iter := q.data.Back(); iter != nil; iter = iter.Prev() {
        fmt.Println("item:", iter.Value)
    }
}
