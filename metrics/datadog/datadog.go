// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package datadog

import (
	"fmt"
	"os"
	"time"

	"github.com/DataDog/chaos-controller/metrics/types"
	chaostypes "github.com/DataDog/chaos-controller/types"
	"github.com/DataDog/datadog-go/statsd"
)

const (
	metricPrefixInjector   = "chaos.injector."
	metricPrefixController = "chaos.controller."
)

// Sink describes a Datadog sink (statsd)
type Sink struct {
	client *statsd.Client
}

// New instantiate a new datadog statsd provider
func New(app types.SinkApp) (*Sink, error) {
	url := os.Getenv("STATSD_URL")

	instance, err := statsd.New(url, statsd.WithTags([]string{"app:" + string(app)}))
	if err != nil {
		return nil, err
	}

	return &Sink{
		client: instance,
	}, nil
}

// Close closes the statsd client
func (d *Sink) Close() error {
	return d.client.Close()
}

// GetSinkName returns the name of the sink
func (d *Sink) GetSinkName() string {
	return string(types.SinkDriverDatadog)
}

// Flush forces the client to send the metrics in the current cache
func (d *Sink) Flush() error {
	return d.client.Flush()
}

// EventWithTags creates a new event with the given title, text and tags and send it
func (d *Sink) EventWithTags(title, text string, tags []string) error {
	e := &statsd.Event{
		Title: title,
		Text:  text,
		Tags:  tags,
	}

	return d.client.Event(e)
}

// MetricInjected increments the injected metric
func (d *Sink) MetricInjected(succeed bool, kind string, tags []string) error {
	status := boolToStatus(succeed)
	t := []string{"status:" + status, "kind:" + kind}
	t = append(t, tags...)

	return d.metricWithStatus(metricPrefixInjector+"injected", t)
}

// MetricCleaned increments the cleaned metric
func (d *Sink) MetricCleaned(succeed bool, kind string, tags []string) error {
	status := boolToStatus(succeed)
	t := []string{"status:" + status, "kind:" + kind}
	t = append(t, tags...)

	return d.metricWithStatus(metricPrefixInjector+"cleaned", t)
}

// MetricReconcile increment reconcile metric
func (d *Sink) MetricReconcile() error {
	return d.metricWithStatus(metricPrefixController+"reconcile", []string{})
}

// MetricReconcileDuration send timing metric for reconcile loop
func (d *Sink) MetricReconcileDuration(duration time.Duration, tags []string) error {
	return d.timing(metricPrefixController+"reconcile.duration", duration, tags)
}

// MetricCleanupDuration send timing metric for cleanup duration
func (d *Sink) MetricCleanupDuration(duration time.Duration, tags []string) error {
	return d.timing(metricPrefixController+"cleanup.duration", duration, tags)
}

// MetricInjectDuration send timing metric for inject duration
func (d *Sink) MetricInjectDuration(duration time.Duration, tags []string) error {
	return d.timing(metricPrefixController+"inject.duration", duration, tags)
}

// MetricDisruptionCompletedDuration sends timing metric for entire disruption duration
func (d *Sink) MetricDisruptionCompletedDuration(duration time.Duration, tags []string) error {
	return d.timing(metricPrefixController+"disruption.completed_duration", duration, tags)
}

// MetricDisruptionOngoingDuration sends timing metric for disruption duration so far
func (d *Sink) MetricDisruptionOngoingDuration(duration time.Duration, tags []string) error {
	return d.timing(metricPrefixController+"disruption.ongoing_duration", duration, tags)
}

// MetricPodsCreated increment pods.created metric
func (d *Sink) MetricPodsCreated(target, instanceName, namespace string, succeed bool) error {
	status := boolToStatus(succeed)
	tags := []string{"target:" + target, "name:" + instanceName, "status:" + status, "namespace:" + namespace}

	return d.metricWithStatus(metricPrefixController+"pods.created", tags)
}

// MetricStuckOnRemoval increments disruptions.stuck_on_removal metric
func (d *Sink) MetricStuckOnRemoval(tags []string) error {
	return d.metricWithStatus(metricPrefixController+"disruptions.stuck_on_removal", tags)
}

// MetricStuckOnRemovalGauge sends disruptions.stuck_on_removal_total metric containing the gauge of stuck disruptions
func (d *Sink) MetricStuckOnRemovalGauge(gauge float64) error {
	return d.client.Gauge(metricPrefixController+"disruptions.stuck_on_removal_total", gauge, []string{}, 1)
}

// MetricDisruptionsGauge sends the disruptions.gauge metric counting ongoing disruptions
func (d *Sink) MetricDisruptionsGauge(gauge float64) error {
	return d.client.Gauge(metricPrefixController+"disruptions.gauge", gauge, []string{}, 1)
}

// MetricDisruptionsCount counts finished disruptions, and tags the disruption kind
func (d *Sink) MetricDisruptionsCount(kind chaostypes.DisruptionKindName, tags []string) error {
	tags = append(tags, fmt.Sprintf("disruption_kind:%s", kind))
	return d.metricWithStatus(metricPrefixController+"disruptions.count", tags)
}

// MetricPodsGauge sends the pods.gauge metric counting existing chaos pods
func (d *Sink) MetricPodsGauge(gauge float64) error {
	return d.client.Gauge(metricPrefixController+"pods.gauge", gauge, []string{}, 1)
}

// MetricRestart sends an increment of the
func (d *Sink) MetricRestart() error {
	return d.metricWithStatus(metricPrefixController+"restart", []string{})
}

// MetricFailedValidation increments the failed validation metric
func (d *Sink) MetricFailedValidation() error {
	return d.metricWithStatus(metricPrefixController+"validation.failed", []string{})
}

func boolToStatus(succeed bool) string {
	var status string
	if succeed {
		status = "succeed"
	} else {
		status = "failed"
	}

	return status
}

func (d *Sink) metricWithStatus(name string, tags []string) error {
	return d.client.Incr(name, tags, 1)
}

func (d *Sink) timing(name string, duration time.Duration, tags []string) error {
	return d.client.Timing(name, duration, tags, 1)
}
