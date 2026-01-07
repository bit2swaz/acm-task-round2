# a/b testing with nginx reverse proxy

this repository contains a containerized environment for conducting a/b testing using nginx as a layer 7 load balancer.

the system routes incoming traffic between two distinct backend service versions using a round-robin algo to ensure an even distribution of requests.

## architecture

the system allows a single entry point (the reverse proxy) to manage traffic flow to multiple backend containers hidden within a private network.

### components

1. **proxy (nginx:alpine)**
* acts as the edge gateway.
* listens on host port `8080` and maps to container port `80`.
* configured with an `upstream` block to manage the pool of backend servers.


2. **app-a (hashicorp/http-echo)**
* simulates the control version of the application.
* returns a static response identifying itself as version a.
* accessible only via the internal docker network.


3. **app-b (hashicorp/http-echo)**
* simulates the variant version of the application.
* returns a static response identifying itself as version b.
* accessible only via the internal docker network.


4. **network (bridge)**
* an isolated internal network (`ab_testing_net`) that allows the proxy to resolve backend hostnames via docker dns.



## prerequisites

* docker engine
* docker compose

## installation and usage

1. clone the repository and navigate to the directory.
2. start the services in detached mode:
```bash
docker-compose up -d

```


3. verify that all three containers (proxy, app-a, app-b) are running:
```bash
docker ps

```



## configuration details

the core logic resides in `nginx.conf`.

### upstream definition

the `upstream` directive defines a group of servers that can be referenced by the `proxy_pass` directive.

```nginx
upstream my_backend_pool {
    server app-a:5678 weight=1;
    server app-b:5678 weight=1;
}

```

* **hostname resolution**: nginx resolves `app-a` and `app-b` using docker's internal dns.
* **load balancing algorithm**: by default, nginx uses round-robin. assigning `weight=1` to both servers ensures a strictly equal 50/50 traffic split. modifying these weights allows for canary deployments (e.g., 90/10 split).

### request routing

the server block listens on port 80 and forwards requests to the upstream pool.

```nginx
location / {
    proxy_pass http://my_backend_pool;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}

```

* **proxy_pass**: passes the http request to the defined upstream group.
* **headers**: original client headers are forwarded to ensure the backend application receives correct metadata about the client IP.

## verification

to verify that traffic is being correctly split between the two containers, execute the following loop in the terminal. this sends 10 sequential requests to the proxy.

```bash
for i in {1..10}; do curl localhost:8080; echo ""; done

```

**expected output:**
the output should alternate between the responses from version a and version b, confirming the round-robin distribution.

```text
🟦 VERSION A

🟩 VERSION B

🟦 VERSION A

🟩 VERSION B
...

```

## design rationale

### why nginx?

nginx was selected for its high performance, event-driven architecture, and robust layer 7 capabilities. unlike a layer 4 load balancer, nginx can inspect http headers and cookies, which is essential for advanced a/b testing scenarios (e.g., sticky sessions via `ip_hash`).

### why docker compose?

docker compose provides an isolated, reproducible environment. it handles internal networking and dns resolution automatically, eliminating the need to manually manage ip addresses or host files.

### why round-robin?

for a stateless a/b test where the goal is to gather unbiased data on two variations, a pure round-robin approach ensures a mathematically even distribution of traffic without the complexity of session persistence.