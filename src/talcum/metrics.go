package talcum

import (
	"fmt"
	"log"
	"strings"
	"time"

	dogstatsd "github.com/DataDog/datadog-go/statsd"
	"gopkg.in/alexcesaro/statsd.v2"
)

// MetricsCollector describes an object capabale of pushing metrics somewhere
type MetricsCollector interface {
	TimeToPickRole(start time.Time)
	RoleChosen(name string)
	RandomRoleChosen()
	RoleError()
	Flush()
}

// MetricsConfig contains options related to Datadog
type MetricsConfig struct {
	Datadog    bool
	StatsdAddr string
	Namespace  string
	Tags       []string
	TagStr     string
}

// DatadogCollector represents a collector that pushes metrics to Datadog
type StatsdCollector struct {
	dc     *dogstatsd.Client
	sc     *statsd.Client
	logger *log.Logger
}

// NewDatadogCollector returns a DatadogCollector using dogstatsd at addr
func NewStatsdCollector(config *MetricsConfig, logger *log.Logger) (*StatsdCollector, error) {
	var err error
	var dc *dogstatsd.Client
	var sc *statsd.Client
	if config.Datadog {
		dc, err = dogstatsd.New(config.StatsdAddr)
		if err != nil {
			return nil, fmt.Errorf("error creating dogstatsd client: %v", err)
		}
		dc.Tags = config.Tags
		dc.Namespace = config.Namespace
	} else {
		ts := []string{}
		for _, t := range config.Tags {
			if strings.Contains(t, "=") {
				ts = append(ts, strings.Split(t, "=")...)
			} else {
				logger.Printf("ignoring malformed metrics tag: %v (expected '<key>=<value>')", t)
			}
		}
		sc, err = statsd.New(statsd.Address(config.StatsdAddr), statsd.TagsFormat(statsd.InfluxDB), statsd.Tags(ts...))
		if err != nil {
			return nil, fmt.Errorf("error creating statsd client: %v", err)
		}
	}
	return &StatsdCollector{
		dc:     dc,
		sc:     sc,
		logger: logger,
	}, nil
}

// TimeToPickRole records the amount of time elapsed picking a role
func (dc *StatsdCollector) TimeToPickRole(start time.Time) {
	if dc != nil {
		dc.timing("time_to_pick_role", time.Since(start))
	}
}

// RoleChosen records the specific role chosen
func (dc *StatsdCollector) RoleChosen(name string) {
	if dc != nil {
		m := fmt.Sprintf("role_assigned.%v", name)
		dc.incr(m)
	}
}

// RoleError increments the error counter
func (dc *StatsdCollector) RoleError() {
	if dc != nil {
		dc.incr("errors")
	}
}

// RandomRoleChosen counts whenever a random role is chosen
func (dc *StatsdCollector) RandomRoleChosen() {
	if dc != nil {
		dc.incr("random_role_chosen")
	}
}

// Flush flushes any pending metrics to statsd
func (dc *StatsdCollector) Flush() {
	if dc != nil && dc.sc != nil {
		dc.sc.Flush()
	}
}

func (dc *StatsdCollector) timing(metric string, duration time.Duration) (err error) {
	defer func() { dc.logger.Printf("metric: timing: %v, %v", metric, duration) }()

	if dc.dc != nil {
		return dc.dc.Timing(metric, duration, nil, 1)
	}
	dc.sc.Timing(metric, duration)
	return nil
}

func (dc *StatsdCollector) incr(metric string) (err error) {
	defer func() { dc.logger.Printf("metric: incr: %v", metric) }()

	if dc.dc != nil {
		return dc.dc.Incr(metric, nil, 1)
	}
	dc.sc.Increment(metric)
	return nil
}
