.PHONY: go go-setup-benchmark

go:
	go run main.go

go-setup-benchmark: wasm
	@count=$(COUNT); \
	mkdir -p test/modules && rm -rf test/modules/*; \
    for i in $$(seq 1 $$count); do \
		cp example/build/release.wasm test/modules/$$i.wasm; \
	done

wasm: $(wildcard ./example/assembly/*.ts)
	cd example && npm run asbuild:release