#!/bin/bash
# This file: /etc/dhcp/dhclient-exit-hooks.d/ddns-gandi.sh

/config/scripts/ddns-gandi -host myrouter -domain example.com -ifname eth0 -apiKey XXXXX >> /var/log/ddns-gandi.log 2>&1

logger ${0##*/}: 'ddns-gandi attempted to update DNS records'
