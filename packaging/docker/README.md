# Surf in Docker

This Compose setup runs Surf's Linux backend and Chromium in a container. Its
persistent identity, browser profile, and files live under
`~/srv/docker/surf`, following the mac-mini host-volume convention. Surf
streams the rendered browser page to the iPad and forwards touch input.

Chromium runs with a normal browser window on an isolated X virtual display.
This keeps the browser compatible with sites that reject Chrome's explicit
headless mode; the virtual display is private to the container and does not
open a window on the Mac desktop.

From the project root, run `make client` once to build and stage the matching
iOS package, then use `make up` to build and start Surf in the background. Use
`make down` to stop it. Browser data and device pairings are preserved.
After startup, `make up` prints clickable links for the public Surf endpoint
and the authenticated local dashboard. Run `make links` to print them again.

Open the Docker administration dashboard with:

```sh
make dashboard
```

The command opens a short-lived, tokenized URL at `http://127.0.0.1:18081`.
The same authenticated URL printed by `make up` is an admin credential; keep it
in your local terminal and do not share it. It expires whenever Surf restarts.
The dashboard port is published only on the Mac loopback interface. It is not
available through the Surf public hostname, the LAN, or Cloudflare.

## Apple Silicon Mac setup

The image builds the backend from this checkout and embeds the verified iOS
package staged by `make client`. The Docker build validates its version,
compatibility generation, size, and SHA-256 before compiling the backend.
Before starting, replace `surf.example.com` in `.env` with your NPM hostname.
The environment file enables Surf's WebSocket tunnel and sets `SURF_DATA_ROOT`
to the mac-mini host-volume path. In Nginx Proxy Manager, route the hostname to
the Mac's LAN IP and port `18080` with HTTPS upstream and WebSocket support
enabled. NPM terminates the outer public HTTPS connection; Surf's paired,
pinned TLS stream stays inside the WebSocket tunnel.

```sh
cp packaging/docker/.env.example packaging/docker/.env
# Edit packaging/docker/.env for your server before starting.
make client
make up
cd packaging/docker
docker compose exec surf surf pair
```

`surf pair` prints a temporary pairing invitation. On the iPad, open Surf,
choose **Pair device**, enter your NPM hostname with port `443` and the
six-digit code, then confirm that the six words match on both devices. Open
the desired page in Surf after pairing.

Keep the Mac awake and connected while using Surf. The container stores the
server identity, browser profile, and downloads in `~/srv/docker/surf`, so
recreating the container does not require pairing again. To inspect the
backend:

```sh
docker compose logs -f surf
docker compose exec surf surf status
```

To stop the backend without deleting its data, run from the project root:

```sh
make down
```

The iOS app requires a rootful jailbreak and is installed from the Surf repo:
`https://seg6.space/surf/`. Surf's compatibility list currently names the
iPad mini 3 through iOS 12.5.7; iOS 12.5.8 is not explicitly listed, so this
client setup is not confirmed for that version.

Surf's backend terminates its pinned TLS itself; it has no plaintext mode.
Keep the backend port `18080` reachable only on the LAN, and expose the public
hostname through the configured WebSocket proxy. See
[Security](../../docs/security.md) for its pairing and host trust model.

The managed uBlock Origin Lite extension is enabled by default. Set
`SURF_CONTENT_BLOCKER=false` temporarily if it interferes with a site's
verification scripts, then run `make up` from the project root to apply it.

## Legacy iPad performance profile

The Docker profile enables adaptive video, preserves the native client
viewport, and caps capture and encoding at 30 FPS. This avoids sending twice
as many H.264 access units as the iPad mini 3 can decode while keeping manga
text sharp. `SURF_SUSPEND_SILENT_AUDIO=true` stops transmitting an all-zero
PCM stream after one second and automatically resumes before the first audio
chunk with a signal.

Client media reports are written to the server log with source, decode,
presentation, drop, queue, and governor fields. Follow them with:

```sh
docker compose logs -f surf | grep 'video: client stats'
```

Raise `STREAM_MAX_FPS` only after the client sustains the current rate without
drops, queue growth, or presentation gaps.

With the iPad connected and the page open, run a repeatable 60-second scrolling
sample from the project root:

```sh
make benchmark
```

The benchmark runs only on demand, shares Surf's network namespace, and does
not expose another port or install tooling in the Surf container. It records
Docker CPU and memory before and after the run plus the server's encode,
input-to-image, client decode, presentation, drop, queue, and governor data.
Reports are written under `.local/benchmarks/`.
