# exporter_proxy

Simple reverse proxy for prometheus exporters.

It is useful when it is difficult to open multiple ports on one server.

## Usage

```
$ exporter_proxy -config config.yml
```

If you make the following settings,

```
listen: "0.0.0.0:9099"

exporters:
  node_exporter:
    path: "/node_exporter/metrics"
    url: "http://localhost:9100/metrics"
  mysqld_exporter:
    path: "/mysqld_exporter/metrics"
    url: "http://localhost:9104/metrics"
```

When you access `http://exporter_proxy_host:9099/node_exporter/metrics`, returns the metrics collected by `node_exporter`.

And of course, `http://exporter_proxy_host:9099/node_exporter/metrics` returns the metrics collected by` mysqld_exporter`.

The part of your `prometheus.yml` is probably as follows.

```
scrape_configs:
  - job_name: "node"
    metrics_path: /node_exporter/metrics
    static_configs:
      - targets: ["exporter_proxy_host:9099"]
  - job_name: "mysqld"
    metrics_path: /mysqld_exporter/metrics
    static_configs:
      - targets: ["exporter_proxy_host:9099"]
```

## Configuration

- Standard example: https://github.com/rrreeeyyy/exporter_proxy/blob/master/config.example.yml
- For docker example: https://github.com/rrreeeyyy/exporter_proxy/blob/master/example/config/config.yml

## Development

### Dependencies

Dependencies are managed by [dep](https://github.com/golang/dep)
