package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

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
	go func() {
		for {
			cmd := exec.Command(`\Users\Harshita\Desktop\go prom\pping`, "-i", "eth0", "-I", "tc", "-F", "json")
			output, err := cmd.Output()
			if err != nil {
				fmt.Println("Error executing command:", err)
				time.Sleep(2 * time.Second)
				continue
			}

			var metrics []MetricData
			err = json.Unmarshal(output, &metrics)
			if err != nil {
				fmt.Println("Error parsing JSON:", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, metric := range metrics {
				labels := prometheus.Labels{
					"src_ip":    metric.SrcIP,
					"dest_ip":   metric.DestIP,
					"src_port":  fmt.Sprintf("%d", metric.SrcPort),
					"dest_port": fmt.Sprintf("%d", metric.DestPort),
					"protocol":  metric.Protocol,
				}
				opsHistogram.With(labels).Observe(float64(len(output)))
			}

			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	opsHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "myapp_ping_output_length",
			Help:    "Length of output from pinging example.com",
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
