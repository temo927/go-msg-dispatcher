# Message Dispatcher

A small Go service that automatically sends messages from a database to a webhook endpoint in batches (2 every 2 minutes), with a REST API to control the scheduler and inspect sent messages.

## Features

- Auto-scheduler: picks **2 queued** messages every **2 minutes**
- On startup: scheduler **auto-starts**
- Status machine: `queued -> processing -> sent` (or `failed` with retries)
- Retries with cap (`MaxRetries`) + last error stored
- (Bonus) Redis cache: stores `messageId` and `sent_at` after successful send
- Swagger/OpenAPI documentation
- Clean architecture (hexagonal), Dockerized

## Quick Start

```bash
# 1) Copy env and adjust values
cp .env.example .env

# IMPORTANT: Edit .env and replace WEBHOOK_URL with YOUR OWN webhook.site UUID.
# Example:
# WEBHOOK_URL=https://webhook.site/<your-uuid-here>

# 2) Build & run (starts API, Postgres, Redis; scheduler auto-starts)
make up

# 3) Seed the database with 4 queued messages (required for a demo run)
make seed

# 4) Tail logs to watch the scheduler pick messages and send them
make logs # # On startup the API immediately drains any queued messages in batch, then processes new ones every 2 minutes. After opening the logs, you may need to wait up to 2 minutes to see the next batch. Once the 4 seeded messages are drained, the scheduler will keep running and log "no queued messages to process."

make sent        # GET  /api/v1/messages/sent   — lists sent messages 

make start       # POST /api/v1/scheduler/start — starts the scheduler (it's already auto-started on boot; this is for manual control)
make stop        # POST /api/v1/scheduler/stop  — stops the scheduler (useful to test stop/start flows)
make redis-dump  # Show cached send metadata in Redis (messageId + sent_at per message)
make swagger     # Prints where the OpenAPI file lives in the repo
make swagger-open # Opens Swagger UI served by the API (http://localhost:8080/swagger/)