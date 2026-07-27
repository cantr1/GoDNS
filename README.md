# GoDNS

GoDNS is a small learning project that pairs a JSON API with a UDP DNS responder. The API manages DNS records in Postgres, and the DNS server answers simple `A` record lookups from those stored records.

The goal is not to build a production DNS server. The goal is to learn the shape of a real Go service: configuration, database access with sqlc, HTTP handlers, long-running UDP service loops, and two concurrent services sharing the same data model.

## What It Does

- Provides an authenticated HTTP API for creating, listing, finding, and deleting DNS records.
- Stores records in Postgres using sqlc-generated query code.
- Runs a DNS server that answers IPv4 `A` record queries.
- Normalizes DNS query names like `kelzlab.archmachine.` into stored names like `kelzlab.archmachine`.
- Converts stored IPv4 string values into DNS wire-format answers.

## Project Shape

```text
cmd/godns/              Application entry point
internal/api/           HTTP API server
internal/auth/          Bearer token helper
internal/database/      sqlc-generated database package
internal/dns/           UDP DNS responder
sql/schema/             Database migrations
sql/queries/            sqlc query definitions
tests/api_endpoints.http API request examples
images/                 Example outputs
```

## Quick Start

Create a `.env` file with the ports, API key, and Postgres connection string:

```env
API_PORT=:8080
DNS_PORT=9090
API_KEY=replace-me
DB_URL=postgres://user:password@localhost:5432/godns?sslmode=disable
DEV_MODE=true
```

Run the application:

```bash
go run ./cmd/godns
```

Seed records through the API using [tests/api_endpoints.http](tests/api_endpoints.http). The examples create records such as:

```json
{
  "name": "kelzlab.archmachine",
  "ttl": 300,
  "class": "IN",
  "type": "A",
  "value": "192.168.1.148"
}
```

Then query the DNS server:

```bash
dig @localhost -p 9090 kelzlab.archmachine A
```

## Example Output

The API returns the records currently stored in Postgres:

![API results showing DNS records returned as JSON](images/api_results.png)

The same records are persisted in the database:

![Postgres table containing seeded DNS records](images/db.png)

Once seeded, the DNS service can answer lookups with the stored IPv4 addresses:

![dig result for kelzlab.archmachine resolving to 192.168.1.148](images/archmachine.png)

![dig result for kelzlab.rhelvm1 resolving to 192.168.1.150](images/rhelmachine.png)

The program logs requests via DNS to the console. When setup as a Systemd service these logs would be accessable via
commands like `journalctl`
![log results](images/logs.png)
## API Endpoints

All API endpoints expect:

```http
Authorization: Bearer <API_KEY>
```

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/health` | Checks that the API is running and the token is valid. |
| `GET` | `/api/records` | Lists all DNS records. |
| `GET` | `/api/records?name=<name>` | Finds records by DNS name. |
| `GET` | `/api/records?value=<value>` | Finds records by record value. |
| `POST` | `/api/records` | Creates a DNS record. |
| `DELETE` | `/api/records?name=<name>` | Deletes records by name. |
| `DELETE` | `/api/records?value=<value>` | Deletes records by value. |
| `DELETE` | `/api/records` | Deletes all records when `DEV_MODE=true`. |

## Current Scope

GoDNS intentionally supports a narrow DNS path right now: simple `A` record lookups backed by the database. That keeps the project focused on understanding DNS request/response flow before adding more protocol features.

Good next learning steps would be:

- Add tests around API handler behavior.
- Return clearer DNS response codes for misses or unsupported query types.
- Add support for `AAAA` or `CNAME` records.
- Replace placeholder setup notes with migration and database bootstrap commands.
