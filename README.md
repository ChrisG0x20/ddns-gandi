# ddns-gandi
Dynamic DNS utility for use with the gandi.net

ddns-gandi is a Dynamic DNS utility for use with the gandi.net domain registrar. I use run it on an Ubiquiti EdgeRouter Lite, a small MIPS-based router. This tool is based on a similar tool by jboelter (https://github.com/jboelter/ddr53), for use with Amazon's Route53.

To use this you'll need a domain name registered with gandi.net and a Gandi API Key. You may generate a Gandi API Key on your user account page.

Install Instructions
---

Modify the shell script as required by your domain configuration. You need to specify your host, domain, the network interface that's connected to your Internet ISP, and your Gandi API Key.

```
    -host myrouter -domain example.com -ifname eth0 -apiKey XXXXX
```
This example would give read your public IP addresses from network interface `eth0`, and create or update your gandi.net DNS records for myrouter.example.com.

Copy the `ddns-gandi` binary and shell script onto your router. (I used `scp` with the EdgeRouter Lite.)

Change the owner/group of each file:
- `chown root:root ddns-gandi`
- `chown root:root ddns-gandi.sh`

Move the shell script into `/etc/dhcp/dhclient-exit-hooks.d/` and the binary into `/config/scripts/`.

To run it manually the first time, execute the shell script. It will produce a log file: `/var/log/ddns-gandi.log`

Build Instructions
---

To build on Linux, requires Go is installed. For a native build:
- `go build -o ddns-gandi`

And, for a MIPS build to run on the EdgeRouter Lite:
- `GOARCH=mips64 go build -o ddns-gandi`
