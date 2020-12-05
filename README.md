# GoCAN

An implementation of a distributed Content Addressable Network using Golang.

## Background

A content addressable network (CAN) is a collection of systems acting as a distributed database where data is addressed based on content, instead of location. This implementation is inspired by [this document](https://people.eecs.berkeley.edu/~sylvia/papers/cans.pdf), which details a method for distributing data across systems using a _d_-dimensional coordinate space. Myself and [TheBigFish](https://github.com/TheBiggerFish) implemented a version of this in C++ using our own packet types during a network theory class in University, so this will be a more refined approach.

## Parameters

_d_ - dimensions \
_r_ - redundancy (backups for data) \
_p_ - listening port \
_join_ - server host:port to join existing CAN

## Methods
## Methods for Clients
| HTTP Method | Description |
| ----------- | ----------- |
| `GET /debug` | Return information about a CAN server, including dimensions, data, and neighbors |
| `POST /trace` | Return server route from entry point to given `key` |
| `PUT /data` | Insert new data into a CAN |
| `PATCH /data` | Update existing data in a CAN |
| `GET /data/{key}` | Retrieve data located at point hashed by `key` |
| `DELETE /data/{key}` | Delete data located at point hashed by `key` |

### Debug Information
**`GET /debug`**

Retrieve all information for a specified CAN server. This data is returned as a JSON object formatted as a `JoinResponse` as found in `/data/types.go`:
```
{
  "dimension": int,
  "redundancy": int,
  "range": {
    "p1": {
      "coords": [
        float64,
        ...
      ]
    },
    "p2": {
      "coords": [
        float64,
        ...
      ]
    }
  },
  "data": {
    "key": "value",
    ...
  },
  "neighbors": {
    "host:port": {
      "p1": {
        "coords": [
          float64,
          ...
        ]
      },
      "p2": {
        "coords": [
          float64,
          ...
        ]
      }
    },
    ...
  }
}
```
### Trace Route
**`POST /trace`**

**HTTP Request:**
| Parameter | Data Type | Description |
| --------- | --------- | ----------- |
| key | string | A string which will be hashed to a coordinate in _d_-dimensional space |

Retrieve a list of servers passed through to reach a point specified by the given `key`. 
## Methods for Servers/Joiners
| HTTP Method | Description |
| ----------- | ----------- |
| `POST /join` | Join a CAN by providing an entry point, listening port, and key |
| `PUT /neighbors` | Add a new neigbor to a CAN server |
| `PATCH /neighbors` | Update an existing neighbor to a CAN server |
| `DELETE /neighbors` | Delete an existing neighbor to a CAN server |



## Roadmap

- Define HTTP content
  - ~~GET/PUT/DELETE/PATCH format for client-server communication~~
  - ~~POST format for server-server communication (and if client wants to join server)~~
- ~~Create network with one server~~
  - ~~Communicate with server, add/retrieve data~~
- Add server to network
  - ~~Assign point in space~~
  - ~~Split region~~
  - ~~Reassign data~~
  - Update neighbor table
  - ~~Create client to receive data and start serving~~
- Client interface with network
  - ~~Client can send requests, server can interpret~~
  - Route requests to appropriate server
  - ~~Route response back to client~~
- Server leaves network
  - Determine neighbor to hand data
  - Reassign region to neighbor
  - Hand data to neighbor
  - Update neighbor table
  - Exit network
- Map network for drawing
  - Figure out flood fill
  - Build a client (maybe a web client)

## Resources

<https://people.eecs.berkeley.edu/~sylvia/papers/cans.pdf> \
<https://gobyexample.com/>
