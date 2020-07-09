# Shooty Game Written in _Go_ For Elastic Path Interview

Using https://ebiten.org/

## Run
`go run main.go`

## Build
`go build main.go`

## WASM
```
go get github.com/hajimehoshi/wasmserve
wasmserve ./
Then access http://localhost:8080/
```

## Compile WASM
`GOOS=js GOARCH=wasm go build -o shooty.wasm github.com/leenattress/goshootygame`

![screenshot](goshooty.gif)