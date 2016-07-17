package updater

import (
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
)

var upd *Updater
var client *fakeMetadataClient

func init() {
	client = &fakeMetadataClient{
		hosts: []metadata.Host{
			{
				Hostname: "Host1",
				AgentIP:  "IP1",
			},
		},
		lock: &sync.Mutex{},
	}

	upd = &Updater{
		rancherHosts:   make(map[string]string),
		MetadataClient: client,
	}
}

type fakeMetadataClient struct {
	hosts []metadata.Host
	lock  *sync.Mutex
}

func (f *fakeMetadataClient) GetHosts() ([]metadata.Host, error) {
	return f.hosts, nil
}

func TestMain(m *testing.M) {
	tmpFile, err := ioutil.TempFile("", "hosts")
	if err != nil {
		log.Fatalf("Error running test, Could not create Temp file [%v]", err)
	}
	_, err = tmpFile.WriteString("127.0.0.1	localhost")
	if err != nil {
		log.Fatalf("Error running test, Could not write init contents into Temp file [%v]", err)
	}
	err = tmpFile.Close()
	if err != nil {
		log.Fatalf("Error running test, Could not close Temp file [%v]", err)
	}
	hostsOrigFile = tmpFile.Name()
	hostsWorkingFile = hostsOrigFile + ".backup"

	defer os.Remove(hostsOrigFile)
	defer os.Remove(hostsWorkingFile)

	upd.Run("")
	os.Exit(m.Run())
}

func TestDetectsHostIpChange(t *testing.T) {
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "IP1",
		},
	}
	upd.Run("")
	hostsMap, err := hostsFileToMap(hostsOrigFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 2 {
		t.Fatalf("Expected 2 entires, found %d", len(hostsMap))
	}
	v, ok := hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "IP1" {
		t.Fatalf("Entry for Host1 not found to be IP1 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "IP2",
		},
	}
	upd.Run("")
	hostsMap, err = hostsFileToMap(hostsOrigFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 2 {
		t.Fatalf("Expected 2 entires, found %d", len(hostsMap))
	}
	v, ok = hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "IP2" {
		t.Fatalf("Entry for Host1 not found to be IP2 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
}

func TestDetectsHostAddition(t *testing.T) {
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "IP1",
		},
	}
	upd.Run("")
	hostsMap, err := hostsFileToMap(hostsOrigFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 2 {
		t.Fatalf("Expected 2 entires, found %d", len(hostsMap))
	}
	v, ok := hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "IP1" {
		t.Fatalf("Entry for Host1 not found to be IP1 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "IP1",
		},
		{
			Hostname: "Host2",
			AgentIP:  "IP2",
		},
	}
	upd.Run("")
	hostsMap, err = hostsFileToMap(hostsOrigFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 3 {
		t.Fatalf("Expected 3 entires, found %d", len(hostsMap))
	}
	v, ok = hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "IP1" {
		t.Fatalf("Entry for Host1 not found to be IP2 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["Host2"]
	if !ok {
		t.Fatalf("Entry for Host2 not found after running updater service with Host2 data")
	}
	if v != "IP2" {
		t.Fatalf("Entry for Host1 not found to be IP2 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
}

func TestDetectsHostDeletion(t *testing.T) {
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "IP1",
		},
		{
			Hostname: "Host2",
			AgentIP:  "IP2",
		},
	}
	upd.Run("")
	hostsMap, err := hostsFileToMap(hostsOrigFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 3 {
		t.Fatalf("Expected 3 entires, found %d", len(hostsMap))
	}
	v, ok := hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "IP1" {
		t.Fatalf("Entry for Host1 not found to be IP2 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["Host2"]
	if !ok {
		t.Fatalf("Entry for Host2 not found after running updater service with Host2 data")
	}
	if v != "IP2" {
		t.Fatalf("Entry for Host1 not found to be IP2 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
	client.lock.Lock()
	client.hosts = []metadata.Host{
		{
			Hostname: "Host1",
			AgentIP:  "IP1",
		},
	}
	upd.Run("")
	hostsMap, err = hostsFileToMap(hostsOrigFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 2 {
		t.Fatalf("Expected 2 entires, found %d", len(hostsMap))
	}
	v, ok = hostsMap["Host1"]
	if !ok {
		t.Fatalf("Entry for Host1 not found after running updater service with Host1 data")
	}
	if v != "IP1" {
		t.Fatalf("Entry for Host1 not found to be IP1 as set, after running updater service with Host1 data")
	}
	v, ok = hostsMap["localhost"]
	if !ok {
		t.Fatalf("Entry for localhost not found after running updater service with localhost data")
	}
	if v != "127.0.0.1" {
		t.Fatalf("Entry for localhost not found to be as set, after running updater service with localhost data")
	}
}

func hostsFileToMap(name string) (map[string]string, error) {
	hostsMap := map[string]string{}
	hostData, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

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
	return hostsMap, nil
}
