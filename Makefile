.PHONY: go wasm

go:
	go run main.go

wasm:
	cd example && npm run asbuild:release