package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    "calypso/octopus"
)

func main() {
    reader := bufio.NewReader(os.Stdin)
    items := &octopus.TransactionStack{}
    for {
        fmt.Print("CLPSO>> ")
        text, _ := reader.ReadString('\n')
        operation := strings.Fields(text)
        switch operation[0] {
        case "BEGIN":
            items.PushTransaction()
        case "ROLLBACK":
            items.RollBackTransaction()
        case "COMMIT":
            items.Commit()
            items.PopTransaction()
        case "END":
            items.PopTransaction()
        case "SET":
            octopus.Set(operation[1], operation[2], items)
        case "GET":
            octopus.Get(operation[1], items)
        case "DELETE":
            octopus.Delete(operation[1], items)
        case "COUNT":
            octopus.Count(operation[1], items)
        case "STOP":
            os.Exit(0)
        default:
            fmt.Println("PANIC: Operation not implemented:", operation[0])
        }
    }
}
