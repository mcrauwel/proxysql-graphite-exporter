# proxysql-graphite-exporter

[![Build Status](https://travis-ci.org/mcrauwel/proxysql-graphite-exporter.svg?branch=master)](https://travis-ci.org/mcrauwel/proxysql-graphite-exporter)

This software will connect to a [ProxySQL](https://github.com/sysown/proxysql) instance, collects metrics and exports these to Graphite.

This check was written by Matthias Crauwels <matthias.crauwels@UGent.be> at Ghent University. It is published with an [MIT license](LICENSE)

## Usage

```
$ ./proxysql-graphite-exporter --help
Usage:
 proxysql-graphite-exporter [OPTIONS]

Application Options:
 -h, --host=     Graphite hostname (default: localhost)
 -p, --port=     Graphite port (default: 2003)
 -P, --protocol= Graphite protocol (default: tcp)
 -d, --dsn=      ProxySQL admin DSN (default: stats:stats@tcp(localhost:6032)/)
 -g, --global    Collect global stats
 -c, --connpool  Collect connection pool stats

Help Options:
 -h, --help      Show this help message
```

## Example

To test this on Linux I run in one terminal

```
vagrant@debian8:~$ ./proxysql-graphite-exporter -g -c
vagrant@debian8:~$
```

while in another terminal I have a netcat socket open that produces following output

```
vagrant@debian8:~$ nc -l 127.0.0.1 -p 2003
proxysql.debian8.global.client_connections_connected-Gauge 0 1503776397
proxysql.debian8.global.active_transactions-Gauge 0 1503776397
proxysql.debian8.global.client_connections_created-Counter 0 1503776397
proxysql.debian8.global.client_connections_aborted-Counter 0 1503776397
proxysql.debian8.global.slow_queries-Counter 0 1503776397
proxysql.debian8.global.proxysql_uptime-Counter 15200 1503776397
proxysql.debian8.global.client_connections_non_idle-Gauge 0 1503776397
proxysql.debian8.global.questions-Counter 0 1503776397
proxysql.debian8.1.127_0_0_1:3307.connection-pool.status-Gauge 2 1503776397
proxysql.debian8.1.127_0_0_1:3307.connection-pool.connused-Gauge 0 1503776397
proxysql.debian8.1.127_0_0_1:3307.connection-pool.connfree-Gauge 0 1503776397
proxysql.debian8.1.127_0_0_1:3307.connection-pool.connerr-Counter 0 1503776397
proxysql.debian8.1.127_0_0_1:3307.connection-pool.queries-Counter 0 1503776397
proxysql.debian8.1.127_0_0_1:3307.connection-pool.bytes_data_sent-Counter 0 1503776397
proxysql.debian8.1.127_0_0_1:3307.connection-pool.bytes_data_recv-Counter 0 1503776397
proxysql.debian8.1.127_0_0_1:3307.connection-pool.latency_us-Gauge 0 1503776397
proxysql.debian8.1.127_0_0_1:3307.connection-pool.connok-Counter 0 1503776397
proxysql.debian8.2.127_0_0_1:3307.connection-pool.status-Gauge 2 1503776397
proxysql.debian8.2.127_0_0_1:3307.connection-pool.connok-Counter 0 1503776397
proxysql.debian8.2.127_0_0_1:3307.connection-pool.queries-Counter 0 1503776397
proxysql.debian8.2.127_0_0_1:3307.connection-pool.bytes_data_recv-Counter 0 1503776397
proxysql.debian8.2.127_0_0_1:3307.connection-pool.latency_us-Gauge 0 1503776397
proxysql.debian8.2.127_0_0_1:3307.connection-pool.connused-Gauge 0 1503776397
proxysql.debian8.2.127_0_0_1:3307.connection-pool.connfree-Gauge 0 1503776397
proxysql.debian8.2.127_0_0_1:3307.connection-pool.connerr-Counter 0 1503776397
proxysql.debian8.2.127_0_0_1:3307.connection-pool.bytes_data_sent-Counter 0 1503776397
```

## Credits

This software collects the metrics as collected by [Percona's Prometheus exporter for ProxySQL](https://github.com/percona/proxysql_exporter).
