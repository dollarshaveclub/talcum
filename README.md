# talcum

Talcum is a tool that allows actors in a distributed system to
coordinate role assignment.

Talcum takes in a configuration file that defines all of the available
roles and the number of actors that should take on the role. If there
are roles available to be filled, an actor will locate a free one and
set a lock on it. If all roles are taken, a random one will be chosen.

## Usage

```
Usage of talcum:
  -app-name string
        the name of the current application (default "app")
  -config-path string
        the path to the role configuration file
  -consul-host string
        the location of Consul (default "localhost:8500")
  -debug
        run in debug mode
  -lock-delay duration
        the delay in between lock attempts (default 100ms)
  -selection-id string
        the ID of the current selection (default "1")
```

## Example configuration file

```
[
  {
    "role_name": "role-1",
    "num": 1
  },
  {
    "role_name": "role-2",
    "num": 2
  },
  {
    "role_name": "role-3",
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
```

The selected role is sent to stdout.
