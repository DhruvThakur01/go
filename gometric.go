package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricData struct {
	Timestamp       int64  `json:"timestamp"`
	SrcIP           string `json:"src_ip"`
	SrcPort         int    `json:"src_port"`
	DestIP          string `json:"dest_ip"`
	DestPort        int    `json:"dest_port"`
	Protocol        string `json:"protocol"`
	FlowEvent       string `json:"flow_event"`
	Reason          string `json:"reason"`
	TriggeredBy     string `json:"triggered_by"`
	RTT             int    `json:"rtt,omitempty"`
	MinRTT          int    `json:"min_rtt,omitempty"`
	SentPackets     int    `json:"sent_packets,omitempty"`
	SentBytes       int    `json:"sent_bytes,omitempty"`
	ReceivedPackets int    `json:"rec_packets,omitempty"`
	ReceivedBytes   int    `json:"rec_bytes,omitempty"`
	MatchOnEgress   bool   `json:"match_on_egress,omitempty"`
}

func recordMetrics() {
	cmd := exec.Command("./pping", "-i", "eth0", "-I", "tc", "-F", "json")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command:", err)
		fmt.Println("Command stderr:", stderr.String())
		return
	}
	output := out.Bytes()

	var metrics []MetricData
	err = json.Unmarshal(output, &metrics)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	for _, metric := range metrics {
		labels := prometheus.Labels{
			"src_ip":    metric.SrcIP,
			"dest_ip":   metric.DestIP,
			"src_port":  strconv.Itoa(metric.SrcPort),
			"dest_port": strconv.Itoa(metric.DestPort),
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
