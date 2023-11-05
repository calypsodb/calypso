package octopus

import "fmt"

func (ts *TransactionStack) PushTransaction() {
    temp := Transaction{Store: make(Map)}
    temp.Next = ts.Top
    ts.Top = &temp
    ts.Size++
}

func (ts *TransactionStack) PopTransaction() {
    if ts.Top == nil {
        fmt.Println("ERROR: No Active Transactions")
    } else {
        ts.Top = ts.Top.Next
        ts.Size--
    }
}

func (ts *TransactionStack) Peek() *Transaction {
    return ts.Top
}

func (ts *TransactionStack) RollBackTransaction() {
    if ts.Top == nil {
        fmt.Println("ERROR: No Active Transaction")
    } else {
        for key := range ts.Top.Store {
            delete(ts.Top.Store, key)
        }
    }
}

func (ts *TransactionStack) Commit() {
    ActiveTransaction := ts.Peek()
    if ActiveTransaction != nil {
        for key, value := range ActiveTransaction.Store {
            GlobalStore[key] = value
            if ActiveTransaction.Next != nil {
                ActiveTransaction.Next.Store[key] = value
            }
        }
    } else {
        fmt.Println("INFO: Nothing to commit")
    }
}

func Get(key string, T *TransactionStack) {
    ActiveTransaction := T.Peek()
    if ActiveTransaction == nil {
        if val, ok := GlobalStore[key]; ok {
            fmt.Println(val)
        } else {
            fmt.Println(key, "not set")
        }
    } else {
        if val, ok := ActiveTransaction.Store[key]; ok {
            fmt.Println(val)
        } else {
            fmt.Println(key, "not set")
        }
    }
}

func Set(key string, value string, T *TransactionStack) {
    ActiveTransaction := T.Peek()
    if ActiveTransaction == nil {
        GlobalStore[key] = value
    } else {
        ActiveTransaction.Store[key] = value
    }
}

func Count(value string, T *TransactionStack) {
    var count int = 0
    ActiveTransaction := T.Peek()
    if ActiveTransaction == nil {
        for _, v := range GlobalStore {
            if v == value {
                count++
            }
        }
    } else {
        for _, v := range ActiveTransaction.Store {
            if v == value {
                count++
            }
        }
    }
    fmt.Println(count)
}

func Delete(key string, T *TransactionStack) {
    ActiveTransaction := T.Peek()
    if ActiveTransaction == nil {
        delete(GlobalStore, key)
    } else {
        delete(ActiveTransaction.Store, key)
    }
    fmt.Println(key, "deleted")
}
