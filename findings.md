This repro shows how Envoy can get stuck with permanently unitialized secrets when
- The secrets are loaded over ADS
- The secret has not changed but the SDS ConfigSource has (in this case, increasing initial_fetch_timeout)

My (incomplete) understanding based on trace/debug logging is:
1. Envoy initialization completes successfully because Envoy is loading all resources for the first time
2. When the initial_fetch_timeout in the DownstreamTlsContext SDS ConfigSource changes, it triggers a Listener update
3. Because Envoy keys secret providers using the [ConfigSource hash](https://github.com/envoyproxy/envoy/blob/ede8074f150e6ede5bdfb4f63853cd2c55c19e63/source/common/secret/secret_manager_impl.h#L88-L89), Envoy creates a new provider/SDS subscription
4. ADS only communicates secret name and most recent version (hash of secret), neither of which have changed
5. From the xDS server perspective, it does not need to respond
6. The new secret provider is stuck "uninitialized" while the old secret provider is destroyed along with the old listener
7. Envoy cannot serve traffic on endpoints that require those secrets until the Envoy process is restarted

This behavior persists even during xDS server restarts, I *think* because Envoy is still communicating its last hash of the
Secret in the ADS request. The only ways I've found to work around this problem are touching the secret (adding whitespace),
or abandoning ADS.

Full trace and debug logs for Envoy and go-control-plane respectively are available, but I've extracted what
appear to me to be the most relevant pieces:

Envoy:

```
[2026-03-26 08:31:04.746][3152732][debug][init] [source/common/init/target_impl.cc:15] init manager Server initializing target LDS
[2026-03-26 08:31:04.753][3152732][debug][config] [source/extensions/config_subscription/grpc/new_grpc_mux_impl.cc:158] Received DeltaDiscoveryResponse for type.googleapis.com/envoy.config.listener.v3.Listener at version 1774539061924
[2026-03-26 08:31:04.754][3152732][debug][init] [source/common/init/manager_impl.cc:24] added shared target SdsApi server_cert to init manager Listener-local-init-manager listener_0 10295629303389728657
[2026-03-26 08:31:04.767][3152732][debug][config] [source/extensions/config_subscription/grpc/new_grpc_mux_impl.cc:158] Received DeltaDiscoveryResponse for type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret at version 1774539061924
[2026-03-26 08:31:04.767][3152732][debug][config] [source/common/tls/server_ssl_socket.cc:80] Secret is updated.
[2026-03-26 08:31:04.768][3152732][debug][init] [source/common/init/watcher_impl.cc:14] shared target SdsApi server_cert initialized, notifying init manager Listener-local-init-manager listener_0 10295629303389728657
...
[2026-03-26 08:31:31.948][3152732][debug][config] [source/extensions/config_subscription/grpc/new_grpc_mux_impl.cc:158] Received DeltaDiscoveryResponse for type.googleapis.com/envoy.config.listener.v3.Listener at version 1774539091941
[2026-03-26 08:31:31.950][3152732][debug][init] [source/common/init/manager_impl.cc:24] added shared target SdsApi server_cert to init manager Listener-local-init-manager listener_0 13130622153538543114
[2026-03-26 08:31:33.949][3152732][warning][config] [source/extensions/config_subscription/grpc/grpc_subscription_impl.cc:130] gRPC config: initial fetch timed out for type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret
[2026-03-26 08:31:33.949][3152732][debug][init] [source/common/init/watcher_impl.cc:14] shared target SdsApi server_cert initialized, notifying init manager Listener-local-init-manager listener_0 13130622153538543114
...
[2026-03-26 08:31:34.951][3152732][debug][secret] [./source/common/secret/secret_manager_impl.h:136] Unregister secret provider. hash key: 1775150063700249849.server_cert
[2026-03-26 08:31:34.951][3152732][debug][init] [source/common/init/target_impl.cc:73] shared target SdsApi server_cert destroyed
[2026-03-26 08:31:34.951][3152732][debug][init] [source/common/init/watcher_impl.cc:31] init manager Listener-local-init-manager listener_0 10295629303389728657 destroyed
[2026-03-26 08:31:34.951][3152732][debug][init] [source/common/init/target_impl.cc:34] target Listener-init-target listener_0 destroyed
```

xDS Server:

```
2026/03/26 08:31:01 setting snapshot for node test-id
2026/03/26 08:31:04 node: test-id, sending delta response for typeURL type.googleapis.com/envoy.config.listener.v3.Listener with resources: [listener_0] removed resources: [] with wildcard: true
2026/03/26 08:31:04 node: test-id, sending delta response for typeURL type.googleapis.com/envoy.config.route.v3.RouteConfiguration with resources: [local_route] removed resources: [] with wildcard: false
2026/03/26 08:31:04 node: test-id, sending delta response for typeURL type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret with resources: [server_cert] removed resources: [] with wildcard: false
2026/03/26 08:31:04 open delta watch ID:2 for type.googleapis.com/envoy.config.listener.v3.Listener Resources:map[] from nodeID: "test-id",  version "1774539061924"
2026/03/26 08:31:04 open delta watch ID:3 for type.googleapis.com/envoy.config.route.v3.RouteConfiguration Resources:map[local_route:{}] from nodeID: "test-id",  version "1774539061924"
2026/03/26 08:31:04 open delta watch ID:4 for type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret Resources:map[server_cert:{}] from nodeID: "test-id",  version "1774539061924"
...
2026/03/26 08:31:31 setting snapshot for node test-id
2026/03/26 08:31:31 node: test-id, sending delta response for typeURL type.googleapis.com/envoy.config.listener.v3.Listener with resources: [listener_0] removed resources: [] with wildcard: true
2026/03/26 08:31:31 open delta watch ID:5 for type.googleapis.com/envoy.config.listener.v3.Listener Resources:map[] from nodeID: "test-id",  version "1774539091941"
EOF
```
