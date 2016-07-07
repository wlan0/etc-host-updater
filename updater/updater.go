package updater

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
)

type Updater struct {
	MetadataClient *metadata.Client
	rancherHosts   map[string]string
}

func (u *Updater) Run(string) {
	if u.rancherHosts == nil {
		u.rancherHosts = make(map[string]string)
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
		}
		hostsMap[host.Hostname] = host.AgentIP
	}

	for rHost := range rancherHosts {
		// a host was deleted
		if _, ok := hostsMap[rHost]; !ok {
			changed = true
		}
	}

	if len(rancherHosts) != len(hostsMap) {
		changed = true
	}

	if !changed {
		return err
	}

	log.Info("change detected in rancher hosts")

	// sycnchronize rancherHosts to be the same as
	// the current view of rancher hosts from metadata service
	for k := range rancherHosts {
		delete(rancherHosts, k)
	}

	for k, v := range hostsMap {
		rancherHosts[k] = v
	}

	hostsWorkingFile := "/etc/hosts.backup"

	if _, err := os.Stat(hostsWorkingFile); os.IsNotExist(err) {
		// This is done to maintain a copy of the original contents
		// of the hosts file
		hostData, err := ioutil.ReadFile("/etc/hosts")
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(hostsWorkingFile, hostData, 0644)
		if err != nil {
			return err
		}
	}
	hostData, err := ioutil.ReadFile(hostsWorkingFile)

	lines := strings.Split(string(hostData), "\n")

	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		ip := ""
		stage2 := false
		hostnames := ""
		for _, val := range line {
			if val == '\t' || val == ' ' {
				stage2 = true
				continue
			}
			if !stage2 {
				ip = ip + string(val)
				continue
			}
			hostnames = hostnames + string(val)

		}
		hostsMap[strings.Trim(hostnames, " ")] = strings.Trim(ip, " ")
	}

	toWrite := ""
	for k, v := range hostsMap {
		toWrite = toWrite + fmt.Sprintf("%s\t%s\n", v, k)
	}

	return ioutil.WriteFile("/etc/hosts", []byte(toWrite), 0644)
}
