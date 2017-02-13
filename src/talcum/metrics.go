package talcum

import (
	"fmt"
	"log"
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

// MetricsCollector describes an object capabale of pushing metrics somewhere
type MetricsCollector interface {
	TimeToPickRole(start time.Time)
	RoleChosen(name string)
	RandomRoleChosen()
	RoleError()
}

// DatadogCollector represents a collector that pushes metrics to Datadog
type DatadogCollector struct {
	c      *statsd.Client
	tags   []string
	logger *log.Logger
}

// NewDatadogCollector returns a DatadogCollector using dogstatsd at addr
func NewDatadogCollector(addr string, namespace string, tags []string, logger *log.Logger) (*DatadogCollector, error) {
	c, err := statsd.New(addr)
	if err != nil {
		return nil, err
	}
	c.Namespace = namespace
	return &DatadogCollector{
		c:      c,
		tags:   tags,
		logger: logger,
	}, nil
}

// TimeToPickRole records the amount of time elapsed picking a role
func (dc *DatadogCollector) TimeToPickRole(start time.Time) {
	if dc != nil {
		dc.timing("time_to_pick_role", time.Since(start), dc.tags)
	}
}

// RoleChosen records the specific role chosen
func (dc *DatadogCollector) RoleChosen(name string) {
	if dc != nil {
		m := fmt.Sprintf("role_assigned.%v", name)
		dc.incr(m, dc.tags)
	}
}

// RoleError increments the error counter
func (dc *DatadogCollector) RoleError() {
	if dc != nil {
		dc.incr("errors", dc.tags)
	}
}

// RandomRoleChosen counts whenever a random role is chosen
func (dc *DatadogCollector) RandomRoleChosen() {
	if dc != nil {
		dc.incr("random_role_chosen", dc.tags)
	}
}

func (dc *DatadogCollector) timing(metric string, duration time.Duration, tags []string) (err error) {
	defer func() { dc.logger.Printf("metric: timing: %v, %v, %v", metric, duration, tags) }()

	return dc.c.Timing(metric, duration, tags, 1)
}

func (dc *DatadogCollector) incr(metric string, tags []string) (err error) {
	defer func() { dc.logger.Printf("metric: incr: %v, %v", metric, tags) }()

	return dc.c.Incr(metric, tags, 1)
}
