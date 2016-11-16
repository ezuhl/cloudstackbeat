package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/ezuhl/cloudstackbeat/config"
	"github.com/xanzy/go-cloudstack/cloudstack"
	//"encoding/json"
)

type Cloudstackbeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Cloudstackbeat{
		done: make(chan struct{}),
		config: config,
	}

	return bt, nil
}

func (bt *Cloudstackbeat) Run(b *beat.Beat) error {
	logp.Info("cloudstackbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	counter := 1
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
		bt.PushDomainLimits(b.Name)
		logp.Info("Event sent")
		counter++
	}
}

func (bt *Cloudstackbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}


func (bt *Cloudstackbeat) PushDomainLimits(beatname string) {
	//
	csClient := cloudstack.NewClient(bt.config.ApiUrl,bt.config.ApiKey,bt.config.ApiSecret, false)
	params   := &cloudstack.ListDomainsParams{}
	params.SetListall(true)


	listDomainResult, err := csClient.Domain.ListDomains(params)
	if err != nil {
		logp.Info("Could not get information from cloudstack management server %s with err %s", bt.config.ApiUrl, err)
		return
	}

	for _,domainObject := range listDomainResult.Domains {
		event := common.MapStr{
			"@timestamp": common.Time(time.Now()),
			"type":       beatname,
			"domain":     domainObject.Name,
			"limits":     domainObject,
		}
		bt.client.PublishEvent(event)
	}



}