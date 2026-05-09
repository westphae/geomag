# geomag Makefile
#
# Convenience wrappers around `go install` and `go test`. The build tag
# `wmm_no_hr` controls whether the high-resolution WMMHR2025 coefficients
# (~530 KB) are embedded in CLI binaries. The default builds bundle both
# WMM and WMMHR so the `--hr` flag works out of the box; the lean target
# omits WMMHR for users with strict binary-size constraints.

.PHONY: install install-lean test test-lean vet vet-lean lint help

help:
	@echo 'geomag — make targets:'
	@echo '  install        install CLIs with full functionality (~12 MB each)'
	@echo '  install-lean   install CLIs without HR support (~11.5 MB each); --hr will error'
	@echo '  test           go test -race -count=1 ./...'
	@echo '  test-lean      go test -tags wmm_no_hr ./...  (verifies the lean build path)'
	@echo '  vet            go vet ./... (both default and lean configs)'
	@echo '  lint           golangci-lint run ./...'

# Default install: full functionality, ~12 MB per CLI binary; `--hr`
# selects the embedded WMMHR2025 model at runtime.
install:
	go install ./cmd/wmm_point ./cmd/wmm_grid ./cmd/wmm_file

# Lean install: omits the embedded WMMHR.COF, ~11.5 MB per binary.
# `--hr` prints a clear "rebuild without -tags wmm_no_hr" error and exits
# non-zero. For storage-conscious specialist users; everyone else should
# use `make install`.
install-lean:
	go install -tags wmm_no_hr ./cmd/wmm_point ./cmd/wmm_grid ./cmd/wmm_file

test:
	go test -race -count=1 ./...

# Verifies that the wmm_no_hr build path compiles and behaves correctly.
test-lean:
	go test -tags wmm_no_hr -count=1 ./...

vet:
	go vet ./...
	go vet -tags wmm_no_hr ./...

vet-lean:
	go vet -tags wmm_no_hr ./...

lint:
	golangci-lint run ./...
