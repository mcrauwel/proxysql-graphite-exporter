package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ShowMax/go-fqdn"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jessevdk/go-flags"
	"github.com/marpaia/graphite-golang"
)

type Options struct {
	GraphiteHost     string `short:"h" long:"host" default:"localhost" description:"Graphite hostname"`
	GraphitePort     int    `short:"p" long:"port" default:"2003" description:"Graphite port"`
	GraphiteProtocol string `short:"P" long:"protocol" default:"tcp" description:"Graphite protocol"`
	ProxySqlDSN      string `short:"d" long:"dsn" default:"stats:stats@tcp(localhost:6032)/" description:"ProxySQL admin DSN"`

	GlobalStats   bool `short:"g" long:"global" description:"Collect global stats"`
	ConnPoolStats bool `short:"c" long:"connpool" description:"Collect connection pool stats"`
}

var options Options
var parser = flags.NewParser(&options, flags.Default)

var proxysql_global_metrics = map[string]string{
	"active_transactions":          "Gauge",
	"client_connections_aborted":   "Counter",
	"client_connections_connected": "Gauge",
	"client_connections_created":   "Counter",
	"client_connections_non_idle":  "Gauge",
	"proxysql_uptime":              "Counter",
	"questions":                    "Counter",
	"slow_queries":                 "Counter",
}

var proxysql_connection_pool_metrics = map[string]string{
	"status":          "Gauge",
	"connused":        "Gauge",
	"connfree":        "Gauge",
	"connok":          "Counter",
	"connerr":         "Counter",
	"queries":         "Counter",
	"bytes_data_sent": "Counter",
	"bytes_data_recv": "Counter",
	"latency_us":      "Gauge",
}

func main() {
	parser.Usage = "[OPTIONS]"
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	graphite_conn := connectToGraphite(options.GraphiteHost, options.GraphitePort, options.GraphiteProtocol, "proxysql")
	proxysql := connectToProxySQL(options.ProxySqlDSN)

	var timestamp = int64(time.Now().Unix())
	//log.Printf("Current timestamp: %d", timestamp)

	if options.GlobalStats {
		globalMetrics := getGlobalStats(proxysql, timestamp)
		graphite_conn.SendMetrics(globalMetrics)
	}

	if options.ConnPoolStats {
		connPoolMetrics := getConnectionPoolStats(proxysql, timestamp)
		graphite_conn.SendMetrics(connPoolMetrics)
	}
}

func getGlobalStats(db *sql.DB, timestamp int64) []graphite.Metric {
	var metrics []graphite.Metric

	data := executeQuery(db, "SELECT Variable_Name, Variable_Value FROM stats_mysql_global")

	m := make(map[string]string)
	for _, row := range data {
		m[row["Variable_Name"]] = row["Variable_Value"]
	}

	for key, value := range m {
		key = strings.ToLower(key)
		value_type, ok := proxysql_global_metrics[key]
		if !ok {
			continue
		}

		graphite_key := fmt.Sprintf("global.%s-%s", key, value_type)
		graphite_value := fmt.Sprintf("%s", value)

		metric := graphite.NewMetric(graphite_key, graphite_value, timestamp)
		metrics = append(metrics, metric)

	}
	return metrics
}

func getConnectionPoolStats(db *sql.DB, timestamp int64) []graphite.Metric {
	var metrics []graphite.Metric

	data := executeQuery(db, "SELECT hostgroup, srv_host, srv_port, * FROM stats_mysql_connection_pool;")

	for _, row := range data {
		var hostgroup, srv_host, srv_port string
		hostgroup, srv_host, srv_port = row["hostgroup"], row["srv_host"], row["srv_port"]

		for key, value := range row {
			switch key {
			case "hostgroup", "srv_host", "srv_port":
				continue
			case "status":
				switch row["status"] {
				case "ONLINE":
					value = "1"
				case "SHUNNED":
					value = "2"
				case "OFFLINE_SOFT":
					value = "3"
				case "OFFLINE_HARD":
					value = "4"
				}
			}

			key = strings.ToLower(key)
			value_type, ok := proxysql_connection_pool_metrics[key]
			if !ok {
				continue
			}

			hostgroup = strings.Replace(hostgroup, ".", "_", -1)
			srv_host = strings.Replace(srv_host, ".", "_", -1)
			srv_port = strings.Replace(srv_port, ".", "_", -1)

			graphite_key := fmt.Sprintf("%s.%s:%s.connection-pool.%s-%s", hostgroup, srv_host, srv_port, key, value_type)
			graphite_value := fmt.Sprintf("%s", value)

			metric := graphite.NewMetric(graphite_key, graphite_value, timestamp)
			metrics = append(metrics, metric)
		}

	}

	return metrics
}

func connectToGraphite(GraphiteHost string, GraphitePort int, GraphiteProtocol string, namespace string) *graphite.Graphite {
	fqdn := strings.Replace(fqdn.Get(), ".", "_", -1)
	graphite_namespace := fmt.Sprintf("%s.%s", namespace, fqdn)

	graphite_conn, err := graphite.GraphiteFactory(GraphiteProtocol, GraphiteHost, GraphitePort, graphite_namespace)
	if err != nil {
		log.Printf("Error opening connection to Graphite: %s", err)
		os.Exit(1)
	}

	return graphite_conn
}

func connectToProxySQL(DSN string) *sql.DB {
	proxysql, err := sql.Open("mysql", options.ProxySqlDSN)
	if err == nil {
		err = proxysql.Ping()
	}

	if err != nil {
		log.Printf("Error opening connection to ProxySQL: %s", err)
		os.Exit(1)
	}

	return proxysql
}

func executeQuery(db *sql.DB, query string) []map[string]string {

	//log.Printf("executeQuery: %s", query)
	rows, err := db.Query(query)
	cols, _ := rows.Columns()

	if err != nil {
		log.Printf("Error: %s", err)
		os.Exit(1)
	}
	defer rows.Close()

	var data []map[string]string

	for rows.Next() {
		pointers := make([]interface{}, len(cols))
		container := make([]string, len(cols))

		for i := range pointers {
			pointers[i] = &container[i]
		}

		rows.Scan(pointers...)

		row := make(map[string]string)
		for i := 0; i < len(cols); i++ {
			row[cols[i]] = container[i]
		}
		data = append(data, row)
	}

	return data
}
