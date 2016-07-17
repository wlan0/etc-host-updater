package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/etc-host-updater/updater"
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
	app.Usage = `Hostname Update Service

	Populates /etc/hosts of a rancher managed container based on currently registered
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
	u := &updater.Updater{
		MetadataClient: metadataClient,
	}

	metadataClient.OnChange(interval, u.Run)
	// It never exits
	return nil
}
