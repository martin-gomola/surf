# CI and build caches

CI follows the supported Docker-only host workflow.

The Docker job runs `make test`, which builds cached Docker stages for the Go
backend and portable C99 client core, then builds the runtime Surf image. The
backend test stage uses Linux AMD64 because Surf's upstream managed-Chromium
fixture is not published for Linux ARM64; the deployed ARM64 image uses Debian
Chromium and is verified by the runtime health checks.

The native-package job remains separate because the jailbroken iOS client
needs the pinned Theos toolchain and iOS 8 SDK. The SDK cache is keyed by its
fetch script, while Buildx caches the build-environment layers under the
`native-buildenv` scope. Native source and package verification still run on
every CI build.

Native desktop clients, cross-platform host archives, and desktop release
publishing are outside this fork's supported workflow.
