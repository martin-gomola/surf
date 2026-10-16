<p align="center">
  <img src="backend/cmd/surf/surf-icon.png" alt="Surf icon" width="80" height="80">
</p>

# Surf for legacy iPads

This fork runs modern websites in Chromium on a Docker host and streams them
to a jailbroken iPhone or iPad. Its primary target is an iPad mini 3 running
iOS 12.5.8, with sharp text and responsive scrolling for manga and ordinary
web browsing.

It is based on [seg6/surf](https://github.com/seg6/surf). This fork supports a
Docker host and the native iOS client. The old desktop host code remains only
to make upstream comparison easier.

## What this fork changes

- Chromium runs headed under Xvfb so sites see a normal browser window.
- No VNC service or VNC port is required.
- Adaptive H.264 streaming keeps the native viewport and reduces frame rate
  before image resolution.
- The default legacy-device ceiling is 30 FPS.
- Silent audio capture pauses automatically and resumes when audio starts.
- A local authenticated dashboard manages pairing, devices, logs, and runtime
  performance.
- `make client` builds the jailbreak package from this checkout. Docker embeds
  that exact package and can offer it to older paired clients.
- CI checks the Docker backend, shared client core, iOS package, secrets, and
  commit privacy.

## Requirements

- A 64-bit macOS, Linux, or Windows host with Docker
- A rootful jailbroken iOS device
- Internet access during the first client and Docker builds
- TCP port `18080` reachable from the iPad for direct LAN use

The iOS package contains armv7 and arm64 slices with minimum versions iOS 6
and iOS 7 respectively. This fork is developed around iOS 12.5.8.

## First setup

```sh
git clone https://github.com/martin-gomola/surf.git
cd surf
cp packaging/docker/.env.example packaging/docker/.env
```

Edit `packaging/docker/.env` before starting. The main settings are:

```dotenv
SURF_PUBLIC_ADDRESS=192.168.1.50:18080
SURF_DATA_ROOT=${HOME}/srv/docker/surf
STREAM_MAX_FPS=30
```

For a reverse proxy or Cloudflare Tunnel, set the public address and tunnel
host to the public hostname and port `443`.

```dotenv
SURF_PUBLIC_ADDRESS=surf.example.net:443
SURF_TUNNEL_HOST=surf.example.net
```

Build the matching iOS client, then start Surf:

```sh
make client
make up
```

`make up` prints the Surf endpoint and a tokenized local dashboard URL.

## Install the iOS client

The verified package is written to `client/ios/packages/`. Install it through
Filza or copy it over SSH:

```sh
scp client/ios/packages/space.seg6.surf_*.deb root@DEVICE_IP:/tmp/surf.deb
ssh root@DEVICE_IP 'dpkg -i /tmp/surf.deb'
ssh root@DEVICE_IP 'su mobile -c uicache || uicache || true'
```

Open Surf on the iPad, run `make dashboard`, and choose **Pair device**.
Compare the six words shown on both devices before approving the pairing.

## Daily commands

| Command | Purpose |
| --- | --- |
| `make up` | Build and start Surf |
| `make down` | Stop Surf without deleting its data |
| `make dashboard` | Open the authenticated local dashboard |
| `make links` | Print the Surf and dashboard URLs |
| `make benchmark` | Run a 60-second scrolling benchmark |
| `make client` | Rebuild, verify, and stage the iOS package |
| `make test` | Run backend and shared client-core tests |

Browser data, pairings, downloads, and the Chromium profile live under
`SURF_DATA_ROOT`. Recreating the container keeps this data. This personal
setup has no backup or restore command; delete the data directory when you
want a completely fresh installation.

## Performance

The dashboard shows the current stream profile, target FPS, iPad decode and
presented FPS, drops, queue depth, governor decision, encode time, and input
latency.

With the iPad connected and a page open, run:

```sh
make benchmark
```

Reports are saved under `.local/benchmarks/` with Docker CPU and memory plus
the client and server media statistics. Raise `STREAM_MAX_FPS` only when the
iPad sustains the current rate without drops or queue growth.

## Limitations

- Widevine is not included, so DRM-protected video may fail.
- Chromium runs inside Docker with `--no-sandbox` because Debian Chromium has
  no usable sandbox in the current container runtime.
- Cloudflare and other site challenges can expire. The persistent Chromium
  profile keeps working sessions, but a site may require verification again.
- The dashboard listens only on `127.0.0.1` and its URL contains a temporary
  admin token. Do not share that URL.
- Native desktop host packages, VNC, headless Chromium, and automatic backups
  are outside this fork's scope.

## Development

```sh
make test
make client-test
make client
```

The public branch must use a GitHub noreply commit address. The Privacy
workflow rejects tracked credentials, absolute home paths, and detected
secrets.

More details:

- [Docker setup](packaging/docker/README.md)
- [Native iOS build](docs/native-build.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Security](docs/security.md)
- [Protocol](docs/protocol.md)
- [Changelog](CHANGELOG.md)

## License

[MIT](LICENSE). Original project by [seg6](https://github.com/seg6).
