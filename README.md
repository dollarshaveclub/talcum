# Talcum

> Talcum allows members of a distributed system to auto-configure themselves.

Talcum takes in a configuration file that defines all of the available
roles and the number of actors that should take on the role. If there
are roles available to be filled, an actor will locate a free one and
set a lock on it. If all roles are taken, a random one will be
chosen. The configuration file can be read in from the Consul KV store
or from disk.

## Usage

```
Usage of talcum:
  -app-name string
    	the name of the current application (default "app")
  -config-path string
    	the path to the role configuration file
  -consul-host string
    	the location of Consul (default "localhost:8500")
  -consul-path string
    	the path to the role configuration in Consul
  -datadog
    	statsd is Datadog (dogstatsd) (default true)
  -debug
    	run in debug mode
  -lock-delay duration
    	the delay in between lock attempts
  -metrics-namespace string
    	Datadog metrics namespace (ignored if not using Datadog) (default "talcum")
  -metrics-tags string
    	Metrics tags (comma-delimited, either datadog <key>:<value> or influxdb <key>=<value> (default "production")
  -selection-id string
    	the ID of the current selection (default "1")
  -statsd-addr string
    	statsd (dogstatsd) address (default "0.0.0.0:8125")
```

## Example configuration

```
[
  {
    "role_name": "role-1",
    "role_definition": "foo,bar,baz",
    "num": 1
  },
  {
    "role_name": "role-2",
    "role_definition": "foo,bar",
    "num": 2
  },
  {
    "role_name": "role-3",
    "role_definition": "alice,bob,mary",
    "num": 3
  }
]
```

This role configuration file defines that one instance of "role1"
should exist, two instances of "role2," etc. All actors involved in
the selection process should have the same configuration file.

## Example run

```
$ talcum -config-path examples/example2.json
  role-2
  foo,bar
```

The selected role definition is written to stdout. The role name is
written to stderr.
