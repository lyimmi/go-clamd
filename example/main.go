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

	if ok, err := c.Scan(ctx, fileName); !ok {
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s has maleware\n", fileName)
	}

	if ok, err := c.ScanAll(ctx, "/tmp"); !ok {
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s has maleware\n", fileName)
	}

	stats, err := c.Stats(ctx)
	if err != nil {
		panic(err)
	}
	d, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(d))
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
