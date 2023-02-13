# pvpc-exporter

Bash script that uploads the PVPC €/MWh data from REData API to influxdb on a daily basis

## Dependencies

- [bash](https://www.gnu.org/software/bash/)
- [coreutils (date)](https://www.gnu.org/software/coreutils/)
- [curl](https://curl.se/)
- [gzip](https://www.gnu.org/software/gzip/)
- [influxdb v2+](https://docs.influxdata.com/influxdb/v2.6/)
- [jq](https://stedolan.github.io/jq/)
- Optional: [make](https://www.gnu.org/software/make/) - for automatic installation support
- [systemd](https://systemd.io/)

## Relevant documentation

- [REData API](https://www.ree.es/en/apidatos)
- [InfluxDB API](https://docs.influxdata.com/influxdb/v2.6/write-data/developer-tools/api/)
- [Systemd Timers](https://www.freedesktop.org/software/systemd/man/systemd.timer.html)

## Installation

### With the Makefile

For convenience, you can install this exporter with the following command or follow the manual process described in the next paragraph.

```
make install
$EDITOR $HOME/.config/pvpc_exporter.conf
```

### Manually

1. Copy `pvpc_exporter.sh` to `$HOME/.local/bin/` and make it executable.

2. Copy `pvpc_exporter.conf` to `$HOME/.config/`, configure it (see the configuration section below) and make it read only.

3. Copy the systemd unit and timer to `$HOME/.config/systemd/user/`:

```
cp pvpc-exporter.* $HOME/.config/systemd/user/
```

5. and run the following command to activate the timer:

```
systemctl --user enable --now pvpc-exporter.timer
```

It's possible to trigger the execution by running manually:

```
systemctl --user start pvpc-exporter.service
```

### Config file

The config file has a few options:

```
INFLUXDB_HOST='influxdb.example.com'
INFLUXDB_API_TOKEN='ZXhhbXBsZXRva2VuZXhhcXdzZGFzZGptcW9kcXdvZGptcXdvZHF3b2RqbXF3ZHFhc2RhCg=='
ORG='home'
BUCKET='pvpc'
```

- `INFLUXDB_HOST` should be the FQDN of the influxdb server.
- `ORG` should be the name of the influxdb organization that contains the pvpc price data bucket defined below.
- `BUCKET` should be the name of the influxdb bucket that will hold the pvpc price data.
- `INFLUXDB_API_TOKEN` should be the influxdb API token value.
  - This token should have write access to the `BUCKET` defined above.

## Exporting prices for dates in the past

If the script is passed a number as the first argument, it will query the REData API for an interval in the past instead of the default date of "today".

This argument should correspond to the number of days in the past relative to the current date.
Take into account that the REData API doesn't allow queries for dates prior to 2014-01-01.

Example:

```
~/.local/bin/pvpc_exporter.sh 30
```

A simple bash loop can be used to query the API with an interval between "today" and 15 days ago:

```
for i in {0..15}; do echo "Exporting data from $(date -I -d "$i days ago")" && ~/.local/bin/pvpc_exporter.sh "$i"; done
```

## Troubleshooting

Run the script manually with bash set to trace:

```
bash -x $HOME/.local/bin/pvpc_exporter.sh
```

Check the systemd service logs and timer info with:

```
journalctl --user --unit pvpc-exporter.service
systemctl --user list-timers
```

## Exported metrics

- price: The pvpc price in €/MWh
- hour: The datetime corresponding to the price above

## Exported metrics example

```
pvpc_price,hour=2023-01-01T23:00:00.000+01:00 price=63.54 1672610400
```

## Uninstallation

### With the Makefile

For convenience, you can uninstall this exporter with the following command or follow the process described in the next paragraph.

```
make uninstall
```

### Manually

Run the following command to deactivate the timer:

```
systemctl --user disable --now pvpc-exporter.timer
```

Delete the following files:

```
~/.local/bin/pvpc_exporter.sh
~/.config/pvpc_exporter.conf
~/.config/systemd/user/pvpc-exporter.timer
~/.config/systemd/user/pvpc-exporter.service
```

## Credits

This project takes inspiration from the following:

- [rare-magma/pbs-exporter](https://github.com/rare-magma/pbs-exporter)
- [mad-ady/prometheus-borg-exporter](https://github.com/mad-ady/prometheus-borg-exporter)
- [OVYA/prometheus-borg-exporter](https://github.com/OVYA/prometheus-borg-exporter)
