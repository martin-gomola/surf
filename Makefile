.PHONY: test backend-test client client-ready client-test client-core-test native-sdk native-buildenv native-package up down links dashboard benchmark

DOCKERFILE := packaging/docker/Dockerfile

test: backend-test client-core-test

backend-test:
	docker build --platform linux/amd64 --target backend-test -f "$(DOCKERFILE)" .

client-core-test:
	docker build --target client-core-test -f "$(DOCKERFILE)" .

client-test: client-core-test

native-sdk:
	native/buildenv/fetch-sdk.sh

native-buildenv:
	docker build -t surf-buildenv native/buildenv

client: native-sdk native-buildenv
	docker run --rm -v "$(CURDIR):/src" surf-buildenv bash -c \
		'make -C /src/client/ios clean package DEBUG=0 && bash /src/client/ios/verify-package.sh'
	@package="$$(tr -d '\r\n' < client/ios/.theos/last_package)"; \
	case "$$package" in /*) ;; *) package="$(CURDIR)/client/ios/$${package#./}" ;; esac; \
	test -f "$$package"; \
	cp "$$package" backend/internal/web/client/client.deb; \
	printf 'Staged custom client: %s\n' "$$package"

native-package: client

client-ready:
	@test -f backend/internal/web/client/client.deb || { \
		echo "Custom iOS client is not staged; run 'make client' first." >&2; exit 1; \
	}

up: client-ready
	cd "$(CURDIR)/packaging/docker" && docker compose up -d --build
	@$(MAKE) --no-print-directory links

down:
	cd "$(CURDIR)/packaging/docker" && docker compose down

links:
	@cd "$(CURDIR)/packaging/docker" && \
		admin=""; attempts=0; \
		while [ -z "$$admin" ] && [ $$attempts -lt 40 ]; do \
			admin="$$(docker compose exec -T surf surf dashboard-url 2>/dev/null)" || true; \
			attempts=$$((attempts + 1)); \
			[ -n "$$admin" ] || sleep 0.25; \
		done; \
		[ -n "$$admin" ] || { echo "Surf started, but dashboard link is not ready." >&2; exit 1; }; \
		server="$$(docker compose exec -T surf sh -c 'case "$$SURF_PUBLIC_ADDRESS" in http://*|https://*) printf "%s" "$$SURF_PUBLIC_ADDRESS" ;; "") printf "https://127.0.0.1:%s" "$$PORT" ;; *) printf "https://%s" "$$SURF_PUBLIC_ADDRESS" ;; esac')"; \
		printf '\nSurf:  %s\nAdmin: %s\n\n' "$$server" "$$admin"

dashboard:
	@url="$$(cd "$(CURDIR)/packaging/docker" && docker compose exec -T surf surf dashboard-url)"; \
	if command -v open >/dev/null 2>&1; then open "$$url"; \
	elif command -v xdg-open >/dev/null 2>&1; then xdg-open "$$url"; \
	elif command -v cmd.exe >/dev/null 2>&1; then cmd.exe /c start "" "$$url"; \
	else printf '%s\n' "$$url"; fi

benchmark:
	packaging/docker/benchmark.sh
