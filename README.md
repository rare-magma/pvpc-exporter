# pvpc-exporter

Bash script that uploads proxmox backup server API info to prometheus' pushgateway.

## Dependencies

- [curl](https://curl.se/)
- [jq](https://stedolan.github.io/jq/)
- Optional: [make](https://www.gnu.org/software/make/) - for automatic installation support

## Relevant documentation

- [Proxmox Backup Server API](https://pvpc.proxmox.com/docs/api-viewer/index.html)
- [Proxmox Backup Server API Tokens](https://pvpc.proxmox.com/docs/user-management.html#api-tokens)
- [Prometheus Pushgateway](https://github.com/prometheus/pushgateway/blob/master/README.md)
- [Systemd Timers](https://www.freedesktop.org/software/systemd/man/systemd.timer.html)

## Installation

<details>
<summary>As normal user</summary>

### With the Makefile

For convenience, you can install this exporter with the following command or follow the manual process described in the next paragraph.

```
make install-user
$EDITOR $HOME/.config/pvpc_exporter.conf
```

### Manually

1. Copy `pvpc_exporter.sh` to `$HOME/.local/bin/` and make it executable.

2. Copy `pvpc_exporter.conf` to `$HOME/.config/`, configure it (see the configuration section below) and make it read only.

3. Edit pvpc-exporter.service and change the following lines:

```
ExecStart=/usr/local/bin/pvpc_exporter.sh
EnvironmentFile=/etc/pvpc_exporter.conf
```

to

```
ExecStart=/home/%u/.local/bin/pvpc_exporter.sh
EnvironmentFile=/home/%u/.config/pvpc_exporter.conf
```

4. Copy the systemd unit and timer to `$HOME/.config/systemd/user/`:

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

</details>
<details>
<summary>As root</summary>

### With the Makefile

For convenience, you can install this exporter with the following command or follow the manual process described in the next paragraph.

```
sudo make install
sudoedit /etc/pvpc_exporter.conf
```

### Manually

1. Copy `pvpc_exporter.sh` to `/usr/local/bin/` and make it executable.

2. Copy `pvpc_exporter.conf` to `/etc/`, configure it (see the configuration section below) and make it read only.

3. Copy the systemd unit and timer to `/etc/systemd/system/`:

```
sudo cp pvpc-exporter.* /etc/systemd/system/
```

4. and run the following command to activate the timer:

```
sudo systemctl enable --now pvpc-exporter.timer
```

It's possible to trigger the execution by running manually:

```
sudo systemctl start pvpc-exporter.service
```

</details>
<br/>

### Config file

The config file has a few options:

```
pvpc_API_TOKEN_NAME='user@pam!prometheus'
pvpc_API_TOKEN='123e4567-e89b-12d3-a456-426614174000'
pvpc_URL='https://pvpc.example.com'
PUSHGATEWAY_URL='https://pushgateway.example.com'
```

- `pvpc_API_TOKEN_NAME` should be the value in the "Token name" column in the Proxmox Backup Server user interface's `Configuration - Access Control - Api Token` page.
- `pvpc_API_TOKEN` should be the value shown when the API Token was created.
  - This token should have at least the `Datastore.Audit` access role assigned to it and the path set to `/datastore`.
- `pvpc_URL` should be the same URL as used to access the Proxmox Backup Server user interface
- `PUSHGATEWAY_URL` should be a valid URL for the [push gateway](https://github.com/prometheus/pushgateway).

### Troubleshooting

Check the systemd service logs and timer info with:

<details>
<summary>As normal user</summary>

```
journalctl --user --unit pvpc-exporter.service
systemctl --user list-timers
```

</details>
<details>
<summary>As root</summary>

```
journalctl --unit pvpc-exporter.service
systemctl list-timers
```

</details>
<br>

## Exported metrics per pvpc store

- pvpc_available: The available bytes of the underlying storage. (-1 on error)
- pvpc_size: The size of the underlying storage in bytes. (-1 on error)
- pvpc_used: The used bytes of the underlying storage. (-1 on error)

## Exported metrics example

```
# HELP pvpc_available The available bytes of the underlying storage. (-1 on error)
# TYPE pvpc_available gauge
# HELP pvpc_size The size of the underlying storage in bytes. (-1 on error)
# TYPE pvpc_size gauge
# HELP pvpc_used The used bytes of the underlying storage. (-1 on error)
# TYPE pvpc_used gauge
pvpc_available {host="pvpc.example.com", store="store1"} -1
pvpc_size {host="pvpc.example.com", store="store1"} -1
pvpc_used {host="pvpc.example.com", store="store1"} -1
# HELP pvpc_available The available bytes of the underlying storage. (-1 on error)
# TYPE pvpc_available gauge
# HELP pvpc_size The size of the underlying storage in bytes. (-1 on error)
# TYPE pvpc_size gauge
# HELP pvpc_used The used bytes of the underlying storage. (-1 on error)
# TYPE pvpc_used gauge
pvpc_available {host="pvpc.example.com", store="store2"} 567317757952
pvpc_size {host="pvpc.example.com", store="store2"} 691587252224
pvpc_used {host="pvpc.example.com", store="store2"} 124269494272
```

## Uninstallation

### With the Makefile

For convenience, you can uninstall this exporter with the following command or follow the process described in the next paragraph.

```
sudo make uninstall
```

### Manually

Run the following command to deactivate the timer:

```
sudo systemctl disable --now pvpc-exporter.timer
```

Delete the following files:

```
/usr/local/bin/pvpc_exporter.sh
/etc/pvpc_exporter.conf
/etc/systemd/system/pvpc-exporter.timer
/etc/systemd/system/pvpc-exporter.service
```

## Credits

This project takes inspiration from the following:

- [mad-ady/borg-exporter](https://github.com/mad-ady/borg-exporter)
- [OVYA/borg-exporter](https://github.com/OVYA/borg-exporter)
