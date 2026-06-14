# GEIS companion on DreamHost (geis.weldedanvil.com)

The companion (`geisd`) aggregates every event feed — including the keyed ones
(AIS ships via AISHub, wildfires via NASA FIRMS) whose secrets must never reach
the client — and serves them as static files. Two roles:

- **cron** (`geisd -refresh`): every ~10 min, fetch + normalize all feeds and
  rewrite `data/payload.json`, `data/satellite.png`, `data/ships.json`.
- **FastCGI** (`dispatch.fcgi` → `geisd -fcgi`): serve those files. No upstream
  calls in the request path, so responses are instant.

## Layout (secrets and data live OUTSIDE the web root)

```
/home/dh_r7wsei/geis/
├── geisd                 # the Linux binary (no cgo)
├── geisd-config.json     # API keys + settings  (NOT web-accessible)
├── data/                 # generated payload.json, satellite.png, ships.json
├── refresh.log           # cron output
└── prod/                 # = web root for geis.weldedanvil.com
    ├── dispatch.fcgi      # FastCGI dispatcher (executable shell wrapper)
    └── .htaccess          # routes all requests to dispatch.fcgi
```

## Build the Linux binary (on your Mac)

DreamHost shared hosting is x86-64 Linux; `geisd` has no cgo, so cross-compile:

```sh
cd <repo>
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o geisd-linux ./cmd/geisd
```

## Deploy (once SSH is set up — awaiting green light)

```sh
HOST=dh_r7wsei@geis.weldedanvil.com          # or the DreamHost SSH host
BASE=/home/dh_r7wsei/geis

# 1. binary + config + dirs (outside web root)
scp geisd-linux $HOST:$BASE/geisd
scp deploy/dreamhost/geisd-config.example.json $HOST:$BASE/geisd-config.json   # then edit in the keys
ssh $HOST "mkdir -p $BASE/data && chmod 600 $BASE/geisd-config.json && chmod +x $BASE/geisd"

# 2. web root: dispatcher + .htaccess
scp deploy/dreamhost/dispatch.fcgi $HOST:$BASE/prod/dispatch.fcgi
scp deploy/dreamhost/.htaccess     $HOST:$BASE/prod/.htaccess
ssh $HOST "chmod +x $BASE/prod/dispatch.fcgi"

# 3. first refresh + cron
ssh $HOST "$BASE/geisd -refresh -config $BASE/geisd-config.json"
ssh $HOST "(crontab -l 2>/dev/null; cat) <<'CRON'
$(cat deploy/dreamhost/crontab.example)
CRON"
```

## Verify

```sh
curl https://geis.weldedanvil.com/healthz          # -> ok
curl https://geis.weldedanvil.com/payload.json     # aggregated JSON
curl -I https://geis.weldedanvil.com/satellite.png # 200 image/png
# Confirm secrets are NOT served:
curl -I https://geis.weldedanvil.com/geisd-config.json  # must be 404
```

Then on the Mac client set `companion.base_url` to `https://geis.weldedanvil.com`
and `companion.enabled` to `true` in the GEIS config.

## Keys (free)

- **AISHub**: register at aishub.net (you must also feed AIS data to use the API);
  put the username in `geisd-config.json`.
- **NASA FIRMS**: request a free MAP_KEY at
  `firms.modaps.eosdis.nasa.gov/api/area/` and put it in `geisd-config.json`.

If FastCGI path routing misbehaves on first deploy, check `error.log` in the
domain's logs dir; the `normalizePath` shim in `geisd` already strips a
`/dispatch.fcgi` prefix, which covers DreamHost's default rewrite.
