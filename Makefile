.PHONY: default build check clean install release test tflint

CLEAN_DAYS=30
INSTALL_DIR=${HOME}/bin
OUT=/tmp/tfctl

default: build

build:
	go build -o $(OUT)

check:
	tools/check.sh --all

clean:
	tools/clean.sh 30

install: build
	mv $(OUT) $(INSTALL_DIR)

release:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make release VERSION=x.y.z"; exit 1; fi
	@if ! echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		echo "Error: VERSION must be a valid semantic version (e.g. v0.2.0) with leading 'v'. Got: $(VERSION)"; \
		exit 1; \
	fi

	git push origin --delete "$(VERSION)" || true
	git tag --delete "$(VERSION)" || true
	git tag "$(VERSION)" --message "Release $(VERSION)."
	git push origin
	git push origin "$(VERSION)"

test: build
	go test ./... --count 1 -v
