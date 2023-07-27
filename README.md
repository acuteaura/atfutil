# atfutil

a simple IPAM tool in go that stores allocations and arbitrary metadata in YAML, allows you to allocate from the smallest fitting block and renders allocations (and optionally the space between allocations) to a markdown table.

## Superblocks

[10.99.0.0/16](example/10.99.0.0-16.md)

## Render all ATF files to Markdown

```
cd example
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

# move to example folder
cd example

# allocate the desired network
./atfutil alloc -d "Proper description" -s 28 -i atf/10.99.0.0-16.atf.yaml -o atf/10.99.0.0-16.atf.yaml 

# render the network setup to the human readable markdown
make
```
