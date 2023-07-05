package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricData struct {
	Timestamp       int64
	SrcIP           string
	SrcPort         int
	DestIP          string
	DestPort        int
	Protocol        string
	FlowEvent       string
	Reason          string
	TriggeredBy     string
	RTT             int
	MinRTT          int
	SentPackets     int
	SentBytes       int
	ReceivedPackets int
	ReceivedBytes   int
	MatchOnEgress   bool
}

func recordMetrics() {
	cmd := exec.Command("./pping", "-i", "eth0", "-I", "tc")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		fmt.Println("Error executing command:", err)
		fmt.Println("Command stderr:", stderr.String())
		return
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Println("Command execution error:", err)
		return
	}

	var metrics []MetricData
	err = json.Unmarshal(out.Bytes(), &metrics)
	if err != nil {
		fmt.Println("Error parsing metrics:", err)
		return
	}

	for _, metric := range metrics {
		labels := prometheus.Labels{
			"src_ip":    metric.SrcIP,
			"dest_ip":   metric.DestIP,
			"src_port":  fmt.Sprintf("%d", metric.SrcPort),
			"dest_port": fmt.Sprintf("%d", metric.DestPort),
			"protocol":  metric.Protocol,
		}
		rttHistogram.With(labels).Observe(float64(metric.RTT))
	}
}

var (
	rttHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "myapp_rtt_vs_timestamp",
			Help:    "RTT (Round Trip Time) vs. Timestamp",
			Buckets: prometheus.LinearBuckets(0, 100, 10),
		},
		[]string{"src_ip", "dest_ip", "src_port", "dest_port", "protocol"},
	)
)

func main() {
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
