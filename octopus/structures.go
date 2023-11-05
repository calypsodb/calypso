package octopus

var GlobalStore = make(map[string]string)

type Map = map[string]string

type Transaction struct {
    Store Map
    Next  *Transaction
}

type TransactionStack struct {
	Top  *Transaction
	Size int
}
