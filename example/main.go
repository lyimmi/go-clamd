package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Lyimmi/go-clamd"
	"os"
)

func main() {
	ctx := context.Background()
	c := clamd.NewClamd()

	ok, _ := c.Ping(ctx)
	if ok {
		fmt.Println("clamd ping ok")
	}

	ok, _ = c.Reload(ctx)
	if ok {
		fmt.Println("clamd reload ok")
	}

	fileName := createTestFile()
	defer os.Remove(fileName)
	ok, _ = c.Scan(ctx, fileName)
	if ok {
		fmt.Printf("%s has no maleware\n", fileName)
	}

	ok, _ = c.ScanAll(ctx, "/tmp")
	if ok {
		fmt.Println("/tmp has no maleware")
	}

	stats, _ := c.Stats(ctx)
	if ok {
		d, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(d))
	}
}

func createTestFile() string {
	f, err := os.CreateTemp("", "go-clamd-test")
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString("this is a test file for go-clamd")
	if err != nil {
		panic(err)
	}
	return f.Name()
}
