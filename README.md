# TR-069 simulator

A CPE simulator for [TR-069](https://en.wikipedia.org/wiki/TR-069) protocol.
It supports both
[TR-098](https://cwmp-data-models.broadband-forum.org/tr-098-1-8-0.html)
and
[TR-181](https://cwmp-data-models.broadband-forum.org/tr-181-2-11-0.html)
datamodel formats.

> [!WARNING]
> **This project is under active development.**
>
> All required TR-069 operations are implemented.
> Current work focuses on optimizations and compatibility.

# Installation

Give it a serial number, point to a datamodel file and the ACS and it should
work out of the box.

See `server/config.go` for available configuration options.

## Docker Compose

Add a service to the `docker-compose.yml` file:
```yaml
services:
  tr069sim:
    image: ghcr.io/localhots/simulatr69:latest
    volumes:
      - /path/to/datamodel.csv:/app/datamodel.csv
      - sim_state:/state
    ports:
      - 7547
    environment:
      SERIAL_NUMBER: G3000E-9799109101
      DATAMODEL_PATH: /app/datamodel.csv
      STATE_PATH: /state/state.json
      ACS_URL: http://myisp/acs
      LOG_LEVEL: debug

volumes:
  sim_state:
```

# Use

All required methods are supported and should behave realistically, except
Upload currently don't do anything. There are quirks though.

## Parameter Normalization

`NORMALIZE_PARAMETERS` when set to `true` will make the simulator attempt to
normalize datamodel parameters. Normalization will bring parameter types to
common `xsd:` prefixed format with values checked against their types: negative
unsigned integer numbers will be corrected to zeroes, boolean "yes" will be
normalized to "true".

Default is `false`.

## Connection Requests

Simulator can accept connection requests made over UDP and HTTP.
If the ACS will set the following parameter values they will be respected:
* `ManagementServer.ConnectionRequestUsername`
* `ManagementServer.ConnectionRequestPassword`

## Firmware Upgrades

The simulator supports firmware upgrades in a simple JSON format:
```json
{"version": "123.45"}
```

If the provided URL can't be loaded or if the firmware file has a different
format the simulator will respond with a fault.

If everything is fine the simulator will change `DeviceInfo.SoftwareVersion`
parameter value in its state and pretend to take time to upgrade and reboot.

# License

[MIT](LICENSE)
