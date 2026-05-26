.PHONY: default build check clean docs install release test tflint

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

docs: recast
	@if [ -z "$(VERSION)" ]; then echo "Usage: make docs VERSION=x.y.z"; exit 1; fi
	@version="$(VERSION)"; version="$${version}"; go run ./tools/docsgen ./docs "$$version"

install: build
	mv $(OUT) $(INSTALL_DIR)

recast:
	@set -e; \
	for cast in docs/asciinema/*.cast; do \
		[ -e "$$cast" ] || continue; \
		gif="$${cast%.cast}.gif"; \
		if [ ! -e "$$gif" ] || [ "$$cast" -nt "$$gif" ]; then \
			echo "=== Rendering $$cast"; \
			agg --verbose "$$cast" "$$gif" --fps-cap 24; \
		else \
			echo "=== Skipping $$cast"; \
		fi; \
	done

release:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make release VERSION=x.y.z"; exit 1; fi

	@if ! echo "$(VERSION)" | grep --extended-regexp --quiet '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		echo "=== ERROR: VERSION must be a valid semantic version (e.g. v0.2.0) with leading 'v'. Got: $(VERSION)"; \
		exit 1; \
	fi

	@if ! grep --extended-regexp --quiet "^\#\# $(VERSION)" CHANGELOG.md; then \
		echo "=== ERROR: CHANGELOG.md entry missing for $(VERSION)."; \
		exit 1; \
	fi

	@if [ -n "$$(git tag --list '$(VERSION)')" ]; then \
		echo "=== ERROR: Local tag $(VERSION) already exists."; \
		exit 1; \
	fi

	@if git ls-remote --tags origin "$(VERSION)" | grep --quiet .; then \
		echo "=== ERROR: Remote tag $(VERSION) already exists."; \
		exit 1; \
	fi

	@if gh release view "$(VERSION)" --repo tfctl/tfctl >/dev/null 2>&1; then \
		echo "=== ERROR: GitHub release $(VERSION) already exists."; \
		exit 1; \
	fi

	@$(MAKE) docs VERSION="$(VERSION)"

	@if [ -n "$$(git status --porcelain -- docs/)" ]; then \
		echo "=== Docs changed after generation; committing docs updates."; \
		git add docs/; \
		git commit --message "docs: regenerate docs for $(VERSION)."; \
	fi

	git push origin --delete "$(VERSION)" || true
	git tag --delete "$(VERSION)" || true
	git tag "$(VERSION)" --message "Release $(VERSION)."
	git push origin
	git push origin "$(VERSION)"

test: build
	go test ./... --count 1 -v
