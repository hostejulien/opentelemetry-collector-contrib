// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package dorisexporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/dorisexporter"

import (
	_ "embed"
	"fmt"

	"go.opentelemetry.io/collector/pdata/pmetric"
)

//go:embed sql/metrics_histogram_ddl.sql
var metricsHistogramDDL string

// dMetricHistogram Histogram Metric to Doris
type dMetricHistogram struct {
	*dMetric               `json:",inline"`
	Timestamp              string         `json:"timestamp"`
	Attributes             map[string]any `json:"attributes"`
	StartTime              string         `json:"start_time"`
	Count                  int64          `json:"count"`
	Sum                    float64        `json:"sum"`
	BucketCounts           []int64        `json:"bucket_counts"`
	ExplicitBounds         []float64      `json:"explicit_bounds"`
	Exemplars              []*dExemplar   `json:"exemplars"`
	Min                    float64        `json:"min"`
	Max                    float64        `json:"max"`
	AggregationTemporality string         `json:"aggregation_temporality"`
}

type metricModelHistogram struct {
	metricModelCommon[dMetricHistogram]
}

func (*metricModelHistogram) metricType() pmetric.MetricType {
	return pmetric.MetricTypeHistogram
}

func (*metricModelHistogram) tableSuffix() string {
	return "_histogram"
}

func (m *metricModelHistogram) add(pm pmetric.Metric, dm *dMetric, e *metricsExporter) error {
	if pm.Type() != pmetric.MetricTypeHistogram {
		return fmt.Errorf("metric type is not histogram: %v", pm.Type().String())
	}

	dataPoints := pm.Histogram().DataPoints()
	for i := 0; i < dataPoints.Len(); i++ {
		dp := dataPoints.At(i)

		exemplars := dp.Exemplars()
		newExemplars := make([]*dExemplar, 0, exemplars.Len())
		for j := 0; j < exemplars.Len(); j++ {
			exemplar := exemplars.At(j)

			newExemplar := &dExemplar{
				FilteredAttributes: exemplar.FilteredAttributes().AsRaw(),
				Timestamp:          e.formatTime(exemplar.Timestamp().AsTime()),
				Value:              e.getExemplarValue(exemplar),
				SpanID:             exemplar.SpanID().String(),
				TraceID:            exemplar.TraceID().String(),
			}

			newExemplars = append(newExemplars, newExemplar)
		}

		bucketCounts := dp.BucketCounts()
		newBucketCounts := make([]int64, 0, bucketCounts.Len())
		for j := 0; j < bucketCounts.Len(); j++ {
			newBucketCounts = append(newBucketCounts, int64(bucketCounts.At(j)))
		}

		explicitBounds := dp.ExplicitBounds()
		newExplicitBounds := make([]float64, 0, explicitBounds.Len())
		for j := 0; j < explicitBounds.Len(); j++ {
			newExplicitBounds = append(newExplicitBounds, explicitBounds.At(j))
		}

		metric := &dMetricHistogram{
			dMetric:                dm,
			Timestamp:              e.formatTime(dp.Timestamp().AsTime()),
			Attributes:             dp.Attributes().AsRaw(),
			StartTime:              e.formatTime(dp.StartTimestamp().AsTime()),
			Count:                  int64(dp.Count()),
			Sum:                    dp.Sum(),
			BucketCounts:           newBucketCounts,
			ExplicitBounds:         newExplicitBounds,
			Exemplars:              newExemplars,
			Min:                    dp.Min(),
			Max:                    dp.Max(),
			AggregationTemporality: pm.Histogram().AggregationTemporality().String(),
		}
		m.data = append(m.data, metric)
	}

	return nil
}
