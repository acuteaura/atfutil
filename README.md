# atfutil

a simple IPAM tool in go, plus my personal network allocations for reference

## Superblocks

[10.99.0.0/16](10.99.0.0-16.md)

## Render all ATF files to Markdown

```
make render
```

Requires golang.

## Compile atfutil

```
go build -o atfutil ./cmd/atfutil/cmd.go
```

## Allocate a new subnet

```bash
# build the latest binary
go build -o atfutil ./cmd/atfutil/cmd.go

# allocate the desired network
./atfutil alloc -d "Proper description" -s 28 -i atf/172.19.0.0.atf.yaml -o atf/172.19.0.0.atf.yaml 

# render the network setup to the human readable markdown
make
```
