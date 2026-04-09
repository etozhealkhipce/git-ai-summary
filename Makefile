GORELEASER_IMAGE ?= goreleaser/goreleaser:v2.15.2

.PHONY: release-snapshot
release-snapshot:
	docker run --rm \
		-v "$$(pwd):/src" \
		-w /src \
		-e CGO_ENABLED=0 \
		$(GORELEASER_IMAGE) \
		release --snapshot --clean
