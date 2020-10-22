# go-match

Match making server written in Go

## Description

This is an HTTP server that's purpose is to group players of varying skill levels
into "matches" with other players of similar skill level.

This server does not facilitate any sort of communication between players. It's
sole purpose is to return a unique id to all clients in the same match to
facilitate communication through some external method.

## Usage

```
Usage of ./go-match:
  -max-size int
        Maximum amount of users for a match (default 32)
  -min-size int
        Minimum amount of users required for a match (default 2)
  -port int
        Port to bind to (default 8080)
  -process-period int
        Amount of time in ms to wait between computing match-ups (default 1000)
  -timeout int
        Amount of time in ms to wait for match (default 30000)
```

## API

### `POST /join` 

Join the match making "lobby" and signal availability to be added to a "match"

#### Request

Body Params:

Content-Type `application/json`

| Field | Type | Description |
|-------|------|-------------|
| score | integer | A number representation of the skill level of the player used to match similarly skilled players. Can be anything but similar skilled players should be numerically close together |

#### Response

Content-Type `application/json`

| Field | Type | Description |
|-------|------|-------------|
| id | string | A UUID that corresponds to the session created. Use this as `session_id` in the `/match` request |

### `GET /match`

This request retrieves the id of a match when one has been created

#### Request

Query Params:

| Field | Type | Description |
|-------|------|-------------|
| session_id | string | A UUID representing the session. This id should be created by using the `/join` API |

#### Response

Content-Type `application/json`

| Field | Type | Description |
|-------|------|-------------|
| id | string | A UUID representing the "match" that has been joined |

