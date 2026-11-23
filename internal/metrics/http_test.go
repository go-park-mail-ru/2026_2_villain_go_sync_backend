package metrics

import (
	"maps"
	"net/http"
	"testing"
	"time"
)

func TestHTTPMetricsObserve(t *testing.T) {
	metrics := NewHTTPMetrics()
	const route = "GET /api/items/{id}"
	metrics.Observe(http.MethodGet, route, http.StatusOK, 125*time.Millisecond)
	metrics.Observe(http.MethodGet, route, http.StatusOK, 375*time.Millisecond)

	families, err := metrics.registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	wantLabels := map[string]string{"method": http.MethodGet, "route": route, "status": "200"}
	const requests = "http_server_requests_total"
	const duration = "http_server_request_duration_seconds"
	seen := make(map[string]bool)
	for _, family := range families {
		name := family.GetName()
		if name != requests && name != duration {
			continue
		}
		seen[name] = true
		series := family.GetMetric()
		if len(series) != 1 {
			t.Fatalf("%s has %d series, want 1", name, len(series))
		}

		metric := series[0]
		labels := make(map[string]string)
		for _, label := range metric.GetLabel() {
			labels[label.GetName()] = label.GetValue()
		}
		if !maps.Equal(labels, wantLabels) {
			t.Errorf("%s labels = %v, want %v", name, labels, wantLabels)
		}

		switch name {
		case requests:
			if count := metric.GetCounter().GetValue(); count != 2 {
				t.Errorf("request count = %v, want 2", count)
			}
		case duration:
			histogram := metric.GetHistogram()
			if count := histogram.GetSampleCount(); count != 2 {
				t.Errorf("duration sample count = %d, want 2", count)
			}
			if sum := histogram.GetSampleSum(); sum != 0.5 {
				t.Errorf("duration sum = %v, want 0.5 seconds", sum)
			}
		}
	}
	for _, name := range []string{requests, duration} {
		if !seen[name] {
			t.Errorf("metric family %q is missing", name)
		}
	}
}
