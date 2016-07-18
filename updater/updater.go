package updater

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
)

var (
	hostsOrigFile = "/etc/hosts"
)

// MetadataClient - This abstraction allows this to be mocked easily in tests
type MetadataClient interface {
	GetHosts() ([]metadata.Host, error)
}

type Updater struct {
	MetadataClient MetadataClient
	rancherHosts   map[string]string
	origData       string
}

func (u *Updater) Run(string) {
	if u.rancherHosts == nil {
		u.rancherHosts = make(map[string]string)
	}
	if u.origData == "" {
		u.origData = `127.0.0.1    localhost
::1    localhost ip6-localhost ip6-loopback
fe00::0    ip6-localnet
ff00::0    ip6-mcastprefix
ff02::1    ip6-allnodes
ff02::2    ip6-allrouters
`

		hostname, err := os.Hostname()
		if err != nil {
			log.Errorf("Error getting hostname of host: %v", err)
			return
		}
		ips, err := net.LookupIP(hostname)
		if err != nil {
			log.Errorf("Error getting IP addresses of host %s, err: %v", hostname, err)
			return
		}
		if len(ips) == 0 {
			log.Errorf("Error getting IP address of host %s, err: No IPs found", hostname)
			return
		}
		u.origData = u.origData + ips[0].String() + "    " + hostname
	}
	err := u.Update(u.rancherHosts)
	if err != nil {
		log.Errorf("Error updating /etc/hosts: [%v]", err)
	}
}

func (u *Updater) Update(rancherHosts map[string]string) error {
	hosts, err := u.MetadataClient.GetHosts()
	if err != nil {
		return err
	}

	changed := false

	hostsMap := map[string]string{}

	for _, host := range hosts {
		if _, ok := hostsMap[host.Hostname]; ok {
			// Do not add subsequent hosts with the
			// duplicate hostnames
			continue
		}
		if ip, ok := rancherHosts[host.Hostname]; !ok || ip != host.AgentIP {
			// If the current host is not a part of the
			// previous set of rancher hosts, then a new host
			// was added
			changed = true
			log.Infof("Adding Host %s %s", host.Hostname, host.AgentIP)
		}
		hostsMap[host.Hostname] = host.AgentIP
	}

	for rHost := range rancherHosts {
		// a host was deleted
		if _, ok := hostsMap[rHost]; !ok {
			log.Infof("Deleting host %s", rHost)
			changed = true
		}
	}

	if len(rancherHosts) != len(hostsMap) {
		changed = true
	}

	if !changed {
		return err
	}

	// sycnchronize rancherHosts to be the same as
	// the current view of rancher hosts from metadata service
	for k := range rancherHosts {
		delete(rancherHosts, k)
	}

	for k, v := range hostsMap {
		rancherHosts[k] = v
	}

	toWrite := u.origData + "\n"

	for k, v := range hostsMap {
		toWrite = toWrite + fmt.Sprintf("%s    %s\n", v, k)
	}

	return ioutil.WriteFile(hostsOrigFile, []byte(toWrite), 0644)
}
