# go-forward

The un-friendly AWS VPN client for linux disables ip forwarding every time it establishes a connection. I suppose this is for "security" but it renders docker containers effectively unable to connect to the internet, which makes them unable to install dependencies.

This very small daemon will use inotify to monitor the `/proc/sys/net/ipv4/ip_forward` "file". If after a write the file does not contain the desired value, it will re-write it. This is written to work for anything that unexpectedly mods your ip_forward configuration, so is not specific to the AWS VPN client.

When run with no argumets, it defaults to enabling forwarding (by writing "1"). You can alter the behavior for your use-case with command line arguments like the following:

```
sudo ./go-forward -value 0
```

This command will check and enforce disabled forwarding whenever the ip_forward file is written to. Be sure to adjust the command in the systemd service file if you intend to run with parameters that are not the default.

# Install

* clone the repo
* `go build .`
* `sudo ln -s /path/to/repo/go-forward /usr/local/bin/go-forward`
* `sudo ln -s /path/to/repo/go-forward.service /etc/systemd/system/go-forward.service`
* `sudo systemctl enable go-forward.service`
* `sudo systemctl start go-forward.service`

# Update

* `git pull`
* `sudo systemctl daemon-reload`
* `sudo systemctl restart go-forward.service`

If you can also just run this program manually, or as a pre-requisit to your AWS VPN client, if you don't want it always running.
