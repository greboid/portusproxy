#### Portus Proxy

Go container designed to be a single point of entry into a standard portus install, that is an instance of portus and a registry.  This saves any complicated setup for your reverse proxy server, just point it at this container.

The following env vars can be setup to configure the proxy, but the defaults are included below and work if you've followed the default setup.

|Name|Default|
|---|---|
|REGISTRY_HOST|registry|
|REGISTRY_PORT|5000|
|PORTUS_HOST|portus|
|PORTUS_PORT|3000|
|PROXY_PORT|8080|