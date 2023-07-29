# go-clamd 
[![Go build & test](https://github.com/Lyimmi/go-clamd/actions/workflows/go-build-test.yml/badge.svg)](https://github.com/Lyimmi/go-clamd/actions/workflows/go-build-test.yml)

A Go client for ClamAV daemon over TCP or UNIX socket.

## Requirements
- Minimum Go version 1.20 
- Only targeted for Linux.

## Features

|     Go     | CalmAv | Description                                                                                                                                                                                                                                                                                                                                         |
|:----------:|:-:|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|    Ping    | PING | Check the server's state. It should reply with "PONG".                                                                                                                                                                                                                                                                                              |
|  Version   | VERSION | Print program and database versions.                                                                                                                                                                                                                                                                                                                |
|   Reload   | RELOAD | Reload the virus databases.                                                                                                                                                                                                                                                                                                                         |
|  Shutdown  | SHUTDOWN | Perform a clean exit.                                                                                                                                                                                                                                                                                                                               |
|    Scan    | SCAN | Scan a file or a directory (recursively) with archive support enabled (if not disabled in clamd.conf). A full path is required.                                                                                                                                                                                                                     |
|  ScanAll   | CONTSCAN | Scan file or directory (recursively) with archive support enabled and don't stop the scanning when a virus is found.                                                                                                                                                                                                                                |
| ScanStream | INSTREAM | Scan a stream of data. The stream is sent to clamd in chunks, after INSTREAM, on the same socket on which the command was sent. This avoids the overhead of establishing new TCP connections and problems with NAT. *Note: do not exceed StreamMaxLength as defined in clamd.conf, otherwise clamd will reply with INSTREAM size limit exceeded and close the connection.* |
|   Stats    | STATS | Replies with statistics about the scan queue, contents of scan queue, and memory usage. The exact reply format is subject to change in future releases.                                                                                                                                                                                             |

## Usage

```shell
go get github.com/lyimmi/go-clamd
```

```golang
import (
    "fmt"
    "github.com/Lyimmi/go-clamd"
)

c := clamd.NewClamd()

if ok, _ := c.Ping(ctx); !ok {
    fmt.Println("clamd nok")
}

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
```

## License
The MIT License (MIT)
