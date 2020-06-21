// build:   GOARCH=mips64 go build -o ddns-gandi
// run:     /config/scripts/ddns-gandi -host myrouter -domain example.com -ifname eth0 -apiKey XXXXX
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

var (
	apiKey string
	domain string
	host   string
	ifname string
	ipv4   string
	ipv6   string
)

func init() {
	flag.StringVar(&apiKey, "apiKey", "", "API key to access gandi account")
	flag.StringVar(&domain, "domain", "", "the domain to update (e.g. example.com)")
	flag.StringVar(&host, "host", "", "the hostname to update (e.g. www)")
	flag.StringVar(&ifname, "ifname", "", "the network interface to update DNS from")

	flag.Parse()
}

func main() {
	log.Println("DynamicDNS Updater")
	if len(apiKey) == 0 || len(domain) == 0 || len(host) == 0 || len(ifname) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	nif, err := net.InterfaceByName(ifname)
	if nil != err {
		log.Printf("failed to locate network interface %v: %v", ifname, err)
		os.Exit(2)
	}

	addrs, err := nif.Addrs()
	if nil != err {
		log.Printf("failed to get network interface addresses for %v: %v", ifname, err)
		os.Exit(2)
	}

	for _, addr := range addrs {
		slices := strings.Split(addr.String(), "/")
		ip := net.ParseIP(slices[0])
                if !ip.IsGlobalUnicast() {
                    continue
                }
		if ip.To4() != nil {
			ipv4 = slices[0]
		} else if 16 == len(ip) {
			ipv6 = slices[0]
		}
	}

	log.Printf("Checking host records for %v.%v are pointed at IPv4: %v and IPv6: %v", host, domain, ipv4, ipv6)

	// gandi.net API calls
	uri := fmt.Sprintf("https://api.gandi.net/v5/livedns/domains/%v/records/%v", domain, host)

	client := &http.Client{}

	req, err := http.NewRequest("GET", uri, nil)
	if nil != err {
		log.Printf("failed to prepare domain records request: %v", err)
		os.Exit(2)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Apikey %v", apiKey))
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if nil != err {
		log.Printf("failed to get domain records: %v", err)
		os.Exit(2)
	}

	if 200 != res.StatusCode {
		body, err := ioutil.ReadAll(res.Body)
		if nil != err {
			log.Printf("failed to read response body: %v", err)
		}

		log.Printf("failed to get domain records: [%v] %v", res.Status, string(body))
		os.Exit(2)
	}

	body, err := ioutil.ReadAll(res.Body)
	if nil != err {
		log.Printf("failed to read domain records: %v", err)
		os.Exit(2)
	}

	log.Printf("Current host records: %v", string(body))

	res.Body.Close()

	type HostRecord struct {
		Rrset_type   string   `json:"rrset_type"`
		Rrset_values []string `json:"rrset_values"`
	}
	var hostRecords []HostRecord
	err = json.Unmarshal(body, &hostRecords)
	if nil != err {
		log.Printf("failed to parse current host records: %v", err)
		os.Exit(2)
	}

	if 0 == len(hostRecords) { // if (the host record doesn't already exist)
		log.Printf("Attempting to create new host records.")

		// create the records
		hrec4 := HostRecord{
			Rrset_type:   "A",
			Rrset_values: []string{ipv4},
		}

		payload, err := json.Marshal(&hrec4)
		if nil != err {
			log.Printf("failed to serialize host record for IPv4: %v", err)
			os.Exit(2)
		}

		req, err := http.NewRequest("POST", uri, bytes.NewBuffer(payload))
		if nil != err {
			log.Printf("failed to prepare IPv4 domain record creation request: %v", err)
			os.Exit(2)
		}

		req.Header.Add("Authorization", fmt.Sprintf("Apikey %v", apiKey))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		res, err := client.Do(req)
		if nil != err {
			log.Printf("failed to create IPv4 domain record: %v", err)
			os.Exit(2)
		}

		body, err := ioutil.ReadAll(res.Body)
		if nil != err {
			log.Printf("failed to read response body: %v", err)
			os.Exit(2)
		}

		log.Printf("Create IPv4 host record result: [%v] %v", res.Status, string(body))

		res.Body.Close()

		if 201 != res.StatusCode {
			log.Printf("failed to create IPv4 domain record")
			os.Exit(2)
		}

		hrec6 := HostRecord{
			Rrset_type:   "AAAA",
			Rrset_values: []string{ipv6},
		}

		payload, err = json.Marshal(&hrec6)
		if nil != err {
			log.Printf("failed to serialize host record for IPv6: %v", err)
			os.Exit(2)
		}

		req, err = http.NewRequest("POST", uri, bytes.NewBuffer(payload))
		if nil != err {
			log.Printf("failed to prepare IPv6 domain record creation request: %v", err)
			os.Exit(2)
		}

		req.Header.Add("Authorization", fmt.Sprintf("Apikey %v", apiKey))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		res, err = client.Do(req)
		if nil != err {
			log.Printf("failed to create IPv6 domain record: %v", err)
			os.Exit(2)
		}

		body, err = ioutil.ReadAll(res.Body)
		if nil != err {
			log.Printf("failed to read response body: %v", err)
			os.Exit(2)
		}

		log.Printf("Create IPv6 host record result: [%v] %v", res.Status, string(body))

		res.Body.Close()

		if 201 != res.StatusCode {
			log.Printf("failed to create IPv6 domain record")
			os.Exit(2)
		}

		log.Printf("Host records created.")
		os.Exit(0)
	}

	// Verify the existing IP addresses
	isUpdateRequired := false
	for _, hr := range hostRecords {
		if ("A" == hr.Rrset_type && ipv4 != hr.Rrset_values[0]) ||
			("AAAA" == hr.Rrset_type && ipv6 != hr.Rrset_values[0]) {
			isUpdateRequired = true
		}
	}

	if !isUpdateRequired {
		log.Printf("No update required.")
		os.Exit(0)
	}

	// Update the existing record
	log.Printf("Host records appear out-of-date. Attempting to update.")

	type UpdateItems struct {
		Items []HostRecord `json:"items"`
	}

	updateItems := UpdateItems{
		Items: []HostRecord{{
			Rrset_type:   "A",
			Rrset_values: []string{ipv4},
		}, {
			Rrset_type:   "AAAA",
			Rrset_values: []string{ipv6},
		},
		},
	}

	payload, err := json.Marshal(&updateItems)
	if nil != err {
		log.Printf("failed to serialize host records for update")
		os.Exit(2)
	}

	req, err = http.NewRequest("PUT", uri, bytes.NewBuffer(payload))
	if nil != err {
		log.Printf("failed to prepare domain records update request: %v", err)
		os.Exit(2)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Apikey %v", apiKey))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err = client.Do(req)
	if nil != err {
		log.Printf("failed to update domain records: %v", err)
		os.Exit(2)
	}

	body, err = ioutil.ReadAll(res.Body)
	if nil != err {
		log.Printf("failed to read response body: %v", err)
		os.Exit(2)
	}

	log.Printf("Update host records result: [%v] %v", res.Status, string(body))

	res.Body.Close()

	if 201 != res.StatusCode {
		log.Printf("Update failed.")
		os.Exit(2)
	}

	log.Printf("Update complete.")
	os.Exit(0)
}
