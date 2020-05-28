package aliyun

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"regexp"
)

type Fetcher struct {
	RegionId        string   `json:"region_id"`
	AccessKeyId     string   `json:"access_key_id"`
	AccessKeySecret string   `json:"access_key_secret"`
	BlackRR         []string `json:"black_rr"`
	OnlyType        []string `json:"only_type"`

	client      *alidns.Client
	blackRRMap  map[string]*regexp.Regexp
	onlyTypeMap map[string]bool
}

func (f *Fetcher) InitFromMap(config map[string]interface{}) (err error) {
	marshal, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(marshal, f)
	if err != nil {
		return err
	}

	return nil
}

func (f *Fetcher) Fetch() (records []string, err error) {
	client, err := alidns.NewClientWithAccessKey(f.RegionId, f.AccessKeyId, f.AccessKeySecret)

	if err != nil {
		return
	}

	f.client = client
	f.blackRRMap = map[string]*regexp.Regexp{}
	f.onlyTypeMap = map[string]bool{}
	for _, rr := range f.BlackRR {
		f.blackRRMap[rr] = regexp.MustCompile(rr)
	}

	for _, t := range f.OnlyType {
		f.onlyTypeMap[t] = true
	}

	domainsRequest := alidns.CreateDescribeDomainsRequest()
	domainsRequest.PageNumber = requests.NewInteger(1)
	domainsRequest.Scheme = "https"
	domainsRequest.PageSize = requests.NewInteger(100)

	domains, err := f.fetchMainDomains(domainsRequest)

	if err != nil {
		return
	}
	var allRecords []alidns.Record

	recordsRequest := alidns.CreateDescribeDomainRecordsRequest()
	recordsRequest.PageNumber = requests.NewInteger(1)
	recordsRequest.PageSize = requests.NewInteger(500)
	for _, domain := range domains {
		recordsRequest.DomainName = domain
		tempRecords, err := f.fetchDomainRecords(recordsRequest)
		if err != nil {
			return []string{}, err
		}
		allRecords = append(allRecords, tempRecords...)
	}

	for _, record := range allRecords {
		if record.Status != "ENABLE" {
			continue
		}

		if len(f.onlyTypeMap) > 0 {
			if _, ok := f.onlyTypeMap[record.Type]; !ok {
				continue
			}
		}

		if !f.isBlack(record.RR) {
			records = append(records, fmt.Sprintf("%s.%s", record.RR, record.DomainName))
		}
	}

	return records, nil
}

func (f *Fetcher) isBlack(rr string) bool {
	if _, ok := f.blackRRMap[rr]; ok {
		return true
	}

	for _, reg := range f.blackRRMap {
		if reg.MatchString(rr) {
			return true
		}
	}

	return false
}

func (f *Fetcher) fetchMainDomains(request *alidns.DescribeDomainsRequest) (domains []string, err error) {
	for {
		response, err := f.client.DescribeDomains(request)
		if err != nil {
			return nil, err
		}

		for _, domain := range response.Domains.Domain {
			domains = append(domains, domain.DomainName)
		}

		if int(response.PageSize) <= len(response.Domains.Domain) {
			value, err := request.PageNumber.GetValue()

			if err != nil {
				return []string{}, err
			}

			request.PageNumber = requests.NewInteger(value + 1)

		} else {
			return domains, nil
		}
	}

}

func (f *Fetcher) fetchDomainRecords(request *alidns.DescribeDomainRecordsRequest) (records []alidns.Record, err error) {
	for {
		response, err := f.client.DescribeDomainRecords(request)
		if err != nil {
			return nil, err
		}

		for _, record := range response.DomainRecords.Record {
			records = append(records, record)
		}

		if int(response.PageSize) <= len(response.DomainRecords.Record) {
			value, err := request.PageNumber.GetValue()

			if err != nil {
				return nil, err
			}

			request.PageNumber = requests.NewInteger(value + 1)

		} else {
			return records, nil
		}
	}
}
