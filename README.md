# GoSubs

Addon de legendas para Stremio com foco em `pt-BR`, traduzindo sob demanda via `GoAI`.

## Fluxo

- procura legenda externa online via Wyzie
- se já existir legenda em português:
  - retorna `Português (Brasil)`
- se não existir, mas houver uma legenda-base traduzível:
  - retorna `🟢 PT-BR (Traduzir)`
- ao clicar:
  - baixa a legenda-base
  - traduz em lotes via `GoAI`
  - salva em cache no disco
  - devolve `.srt`

## Endpoints

- `GET /healthz`
- `GET /manifest.json`
- `GET /subtitles/:type/*tail`
- `GET /subtitles/source/:token`
- `GET /subtitles/generate/:token`
- `GET /subtitles/cache/:key`
- `POST /admin/subtitles/cache/cleanup`

## Environment

- `GOSUBS_ADMIN_API_KEY`
- `GOSUBS_TOKEN_SECRET` (opcional; fallback para `GOSUBS_ADMIN_API_KEY`)
- `GOSUBS_WYZIE_API_KEY`
- `GOSUBS_WYZIE_BASE_URL` (default `https://sub.wyzie.io`)
- `GOSUBS_WYZIE_SOURCE` (default `all`)
- `GOSUBS_GOAI_BASE_URL`
- `GOSUBS_GOAI_API_KEY`
- `GOSUBS_CACHE_DIR` (default `${TMPDIR}/gosubs-cache`)
- `GOSUBS_CACHE_TTL` (default `48h`)
- `GOSUBS_MAX_BATCH_CHARS` (default `4000`)
- `GOSUBS_HTTP_ADDR` ou `PORT`

## Deploy

- hostname publico: `gosubs.duckdns.org`
- namespace k3s: `gosubs`
- o workflow usa o mesmo padrao de `k3s` dos outros projetos:
  - build/push GHCR
  - copia `deploy/k8s/gosubs`
  - recria o Secret `gosubs-env`
  - `kubectl apply -k`

## Run

```bash
export GOSUBS_ADMIN_API_KEY=devsecret
export GOSUBS_WYZIE_API_KEY=wyzie-...
export GOSUBS_GOAI_BASE_URL=http://localhost:8088
export GOSUBS_GOAI_API_KEY=devsecret
go run ./cmd/gosubs
```
