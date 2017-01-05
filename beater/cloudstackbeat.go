package beater

import (
	"fmt"
	"time"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"
	"github.com/xanzy/go-cloudstack/cloudstack"
	"reflect"
	"strconv"
)

type Config struct {
	Period time.Duration `config:"period"`
	ApiKey string `config:"cloudstackkey"`
	ApiSecret string `config:"cloudstacksecret"`
	ApiUrl string `config:"cloudstackurl"`
}



type Cloudstackbeat struct {
	done   chan struct{}
	config Config
	client publisher.Client
}
type ElasticDomain struct {
	Cpuavailable              int64
	Cpulimit                  int64
	Cputotal                  int64
	Haschild                  bool
	Id                        string
	Ipavailable               int64
	Iplimit                   int64
	Iptotal                   int64
	Level                     int64
	Memoryavailable           int64
	Memorylimit               int64
	Memorytotal               int64
	Name                      string
	Networkavailable          int64
	Networkdomain             string
	Networklimit              int64
	Networktotal              int64
	Parentdomainid            string
	Parentdomainname          string
	Path                      string
	Primarystorageavailable   int64
	Primarystoragelimit       int64
	Primarystoragetotal       int64
	Projectavailable          int64
	Projectlimit              int64
	Projecttotal              int64
	Secondarystorageavailable int64
	Secondarystoragelimit     int64
	Secondarystoragetotal     int64
	Snapshotavailable         int64
	Snapshotlimit             int64
	Snapshottotal             int64
	State                     string
	Templateavailable         int64
	Templatelimit             int64
	Templatetotal             int64
	Vmavailable               int64
	Vmlimit                   int64
	Vmtotal                   int64
	Volumeavailable           int64
	Volumelimit               int64
	Volumetotal               int64
	Vpcavailable              int64
	Vpclimit                  int64
	Vpctotal                  int64
}
// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := Config{}
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
		logp.Warn("Could not get information from cloudstack management server %s with err %s", bt.config.ApiUrl, err)
		return
	}

	for _,domainObject := range listDomainResult.Domains {

		event := common.MapStr{
			"@timestamp": common.Time(time.Now()),
			"type":       beatname,
			"domain":     domainObject.Name,
			"limits":     bt.getElasticDomain(domainObject),
		}
		bt.client.PublishEvent(event)
	}



}




func (bt *Cloudstackbeat) getElasticDomain(cloudstackDomain *cloudstack.Domain) ElasticDomain {

	correctDataDomain := ElasticDomain{}
	voc := reflect.ValueOf(*cloudstackDomain)

	toe := reflect.TypeOf(correctDataDomain)
	voe := reflect.ValueOf(&correctDataDomain)

	elm := voe.Elem()
	for i := 0; i < toe.NumField(); i += 1 {
		name := toe.Field(i).Name
		cValue := voc.FieldByName(name)
		//fmt.Printf("name: %s   value: %s  type: %s \n",name, cValue.String(),cValue.Type().String())
		if(cValue.Type().String() == "string") {
			//is string

			if (elm.Field(i).Kind() == reflect.String) {
				elm.FieldByName(name).SetString(cValue.String())
				continue
			}else if (elm.Field(i).Kind() == reflect.Int64) {
				if (cValue.String() == "Unlimited") {
					elm.FieldByName(name).SetInt(-1)
					continue
				}
				newInteger, err := strconv.ParseInt(cValue.String(), 10, 64); if (err == nil) {
					elm.FieldByName(name).SetInt(newInteger)
					continue
				}
			} else if (elm.Field(i).Kind() == reflect.Bool) {
				boolValue,err := strconv.ParseBool(cValue.String()); if(err == nil){
					elm.FieldByName(name).SetBool(boolValue)
					continue
				}
			}
		}else if(cValue.Type().String() == "bool"){
			elm.FieldByName(name).SetBool(cValue.Bool())

		}else if(cValue.Type().String() == "int" || cValue.Type().String() == "int32"  || cValue.Type().String() == "int64" ){
			elm.FieldByName(name).SetInt(cValue.Int())

		}
	}
	return correctDataDomain
}
//