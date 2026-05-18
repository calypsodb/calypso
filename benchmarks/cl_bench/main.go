package main

import (
	"fmt"
	"log"
	"time"

	"github.com/calypsodb/calypso"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {
	db, err := calypso.Open("./testdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	startInsert := time.Now()
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		value := []byte(fmt.Sprintf("value%d", i))
		if err := db.Put(key, value); err != nil {
			log.Fatal(err)
		}
	}
	insertTime := time.Since(startInsert)
	fmt.Printf("Insertion time for 1000 records: %s\n", insertTime)

	startRead := time.Now()
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("key%d", i))
		value, err := db.Get(key)
		if err != nil {
			log.Fatal(err)
		}
		_ = value
	}
	readTime := time.Since(startRead)
	fmt.Printf("Read time for 1000 records: %s\n", readTime)

	p := plot.New()
	p.Title.Text = "Insert and Read Time"
	p.Y.Label.Text = "Time (s)"

	pts := make(plotter.XYs, 2)
	pts[0].X = 0
	pts[0].Y = insertTime.Seconds()
	pts[1].X = 1
	pts[1].Y = readTime.Seconds()

	scatter, err := plotter.NewScatter(pts)
	if err != nil {
		log.Panic(err)
	}
	p.Add(scatter)

	if err := p.Save(5*vg.Inch, 3*vg.Inch, "scatter.png"); err != nil {
		log.Fatal(err)
	}
}
