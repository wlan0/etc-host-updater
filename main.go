package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/go-rancher-metadata/metadata"
)

const (
	VERSION     = "0.0.1"
	metadataURL = "http://rancher-metadata/2015-12-19"
)

func main() {
	exit := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	app := cli.NewApp()
	app.Author = "Rancher Labs"
	app.EnableBashCompletion = true
	app.Version = VERSION
	app.Usage = `Kubernetes Hostname Service

	Populates /etc/hosts of the kube-apiserver container based on currently registered
	hosts in a given rancher environment`
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "update-interval",
			Value: 5,
			Usage: "time interval between refreshes of host list (in seconds)",
		},
	}
	app.Action = func(c *cli.Context) {
		exit(run(c))
	}

	exit(app.Run(os.Args))
}

func run(c *cli.Context) error {
	metadataClient, err := metadata.NewClientAndWait(metadataURL)
	if err != nil {
		return err
	}

	interval := c.Int("update-interval")
	rancherHosts := map[string]string{}

	for {
		err = repeat(metadataClient, rancherHosts)
		if err != nil {
			log.Errorf("Error updating /etc/hosts with hosts: [%v]", err)
		}
		<-time.After(time.Duration(time.Duration(interval) * time.Second))
	}
}

func repeat(metadataClient *metadata.Client, rancherHosts map[string]string) error {
	hosts, err := metadataClient.GetHosts()
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

	// if a host was deleted, this will be true
	if len(rancherHosts) != len(hostsMap) {
		changed = true
	}

	if !changed {
		log.Info("No change detected in rancher hosts")
		return nil
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
		fd, err := os.Open("/etc/hosts")
		if err != nil {
			return err
		}
		defer fd.Close()

		hostData, err := ioutil.ReadAll(fd)
		if err != nil {
			return err
		}
		fd.Close()

		err = ioutil.WriteFile(hostsWorkingFile, hostData, 0644)
		if err != nil {
			return err
		}
	}

	hostsFile, err := os.OpenFile(hostsWorkingFile, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer hostsFile.Close()

	hostData, err := ioutil.ReadAll(hostsFile)

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
