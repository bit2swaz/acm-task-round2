# weighted a/b testing proxy

a containerized a/b testing environment that uses a **custom reverse proxy written in go** to manage traffic distribution.

instead of relying on standard nginx configuration, this version of the project implements the routing logic manually at the application layer using go's `net/http` standard library.

## architecture

* **proxy (custom go)**: a lightweight layer 7 load balancer built from scratch. it implements a weighted random algorithm to split traffic between backends.
* **backends**: two isolated containers (`app-a`, `app-b`) running `hashicorp/http-echo` to simulate different application versions.
* **infrastructure**: orchestrated via docker compose on an internal bridge network.

## why manual go implementation?

i chose to write the proxy logic manually rather than using `httputil.ReverseProxy` or nginx because:
1.  **granular control**: explicitly handling the request/response lifecycle (copying headers, streaming body bytes) allows for deeper inspection and modification.
2.  **custom algorithms**: implementing the weighted selection logic directly in code (`rand.Intn`) demonstrates how load balancers make decisions at a low level.
3.  **observability**: the proxy logs every routing decision to stdout, providing immediate insight into traffic flow.

## implementation details

the core logic (`main.go`) avoids external dependencies.

* **algorithm**: calculates a cumulative weight threshold to select targets.
* **forwarding**:
    * constructs a new upstream request.
    * copies original client headers to preserve context.
    * adds `X-Forwarded-By` for tracing.
    * streams the response body directly to the client using `io.Copy`.

## verification

### 1. traffic distribution
running a loop of 10 requests shows the weighted distribution in action:

```bash
$ for i in {1..10}; do curl localhost:8080; echo ""; done

🟩 VERSION B (Weight 5)
🟩 VERSION B (Weight 5)
🟦 VERSION A (Weight 5)
🟩 VERSION B (Weight 5)
🟩 VERSION B (Weight 5)
🟦 VERSION A (Weight 5)
🟦 VERSION A (Weight 5)
🟩 VERSION B (Weight 5)
🟩 VERSION B (Weight 5)
🟦 VERSION A (Weight 5)
```

### 2. routing logs

the proxy outputs its decision-making process in real-time:

```text
2026/01/11 05:58:34 [Proxy] Routing 172.20.0.1:38550 -> http://app-a:5678
2026/01/11 05:58:34 [Proxy] Routing 172.20.0.1:38552 -> http://app-b:5678
2026/01/11 05:58:34 [Proxy] Routing 172.20.0.1:38568 -> http://app-b:5678

```

## how to run

```bash
# build the go proxy and start services
docker-compose up --build -d
```