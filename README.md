# go-clamav
A GO client to connect to ClamAV's daemon (clamd) over TCP, or UNIX socket.

## Requirements
- ClamAV and its daemon have to be installed on the host machine.
- Minimum GO version 1.20 

## Features

| GO | CalmAv | Description | Status |
|--|--|--|--|
| Ping | PING | Check the server's state. It should reply with "PONG". | done |
| Version | VERSION | Print program and database versions. | done |
| Reload | RELOAD | Reload the virus databases. | done |
| Scan | SCAN | Scan a file or a directory (recursively) with archive support enabled (if not disabled in clamd.conf). A full path is required. | done |
| ScanAll | CONTSCAN | Scan file or directory (recursively) with archive support enabled and don't stop the scanning when a virus is found. | done |
| ScanStream | INSTREAM | Scan a stream of data. The stream is sent to clamd in chunks, after INSTREAM, on the same socket on which the command was sent. This avoids the overhead of establishing new TCP connections and problems with NAT. *Note: do not exceed StreamMaxLength as defined in clamd.conf, otherwise clamd will reply with INSTREAM size limit exceeded and close the connection.*  | todo |

## Usage
TODO: create examples

## License
The MIT License (MIT)
