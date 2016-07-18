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
	upd.origData = "127.0.0.1    localhost localhost-ip4"
	hostsOrigFile = tmpFile.Name()

	defer os.Remove(hostsOrigFile)

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
	hostsMap, err := parseHostsOrigFile(hostsOrigFile)
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
	hostsMap, err = parseHostsOrigFile(hostsOrigFile)
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
	hostsMap, err := parseHostsOrigFile(hostsOrigFile)
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
	hostsMap, err = parseHostsOrigFile(hostsOrigFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 4 {
		t.Fatalf("Expected 4 entires, found %d", len(hostsMap))
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
	hostsMap, err := parseHostsOrigFile(hostsOrigFile)
	client.lock.Unlock()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(hostsMap) != 4 {
		t.Fatalf("Expected 4 entires, found %d", len(hostsMap))
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
	hostsMap, err = parseHostsOrigFile(hostsOrigFile)
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

func parseHostsOrigFile(file string) (map[string]string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	hostsMap := map[string]string{}
	lines := string(data)
	for _, line := range strings.Split(lines, "\n") {
		elements := strings.Split(line, "    ")
		if len(elements) < 2 {
			continue
		}
		names := strings.Split(strings.Trim(elements[1], " "), " ")
		for _, name := range names {
			hostsMap[name] = elements[0]
		}
	}
	return hostsMap, nil
}
