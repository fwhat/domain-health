package aliyun

import (
	"fmt"
	"github.com/Dowte/domain-health/common"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"regexp"
)

var log = common.Log

type Fetcher struct {
	RegionId        string
	AccessKeyId     string
	AccessKeySecret string
	BlackRR         []string

	client     *alidns.Client
	blackRRMap map[string]*regexp.Regexp
}

type recordStruct struct {
	rr     string
	domain string
}

func (f *Fetcher) Fetch() (records []string, err error) {
	client, err := alidns.NewClientWithAccessKey(f.RegionId, f.AccessKeyId, f.AccessKeySecret)

	if err != nil {
		return
	}

	f.client = client
	f.blackRRMap = map[string]*regexp.Regexp{}
	for _, rr := range f.BlackRR {
		f.blackRRMap[rr] = regexp.MustCompile(rr)
	}

	domainsRequest := alidns.CreateDescribeDomainsRequest()
	domainsRequest.PageNumber = requests.NewInteger(1)
	domainsRequest.Scheme = "https"
	domainsRequest.PageSize = requests.NewInteger(100)

	domains, err := f.fetchMainDomains(domainsRequest, []string{})

	if err != nil {
		return
	}
	var allRecords []recordStruct

	recordsRequest := alidns.CreateDescribeDomainRecordsRequest()
	recordsRequest.PageNumber = requests.NewInteger(1)
	recordsRequest.PageSize = requests.NewInteger(500)
	for _, domain := range domains {
		recordsRequest.DomainName = domain
		tempRecords, err := f.fetchDomainRecords(recordsRequest, []recordStruct{})
		if err != nil {
			return []string{}, err
		}
		allRecords = append(allRecords, tempRecords...)
	}

	for _, record := range allRecords {
		if !f.isBlack(record.rr) {
			records = append(records, fmt.Sprintf("%s.%s", record.rr, record.domain))
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

func (f *Fetcher) fetchMainDomains(request *alidns.DescribeDomainsRequest, tempDomains []string) (domains []string, err error) {
	response, err := f.client.DescribeDomains(request)
	if err != nil {
		return
	}

	for _, domain := range response.Domains.Domain {
		tempDomains = append(tempDomains, domain.DomainName)
	}

	if int(response.TotalCount) > len(tempDomains) {
		value, err := request.PageNumber.GetValue()

		if err != nil {
			return []string{}, err
		}

		request.PageNumber = requests.NewInteger(value + 1)

		return f.fetchMainDomains(request, tempDomains)
	}

	return tempDomains, nil
}

func (f *Fetcher) fetchDomainRecords(request *alidns.DescribeDomainRecordsRequest, tempRecords []recordStruct) (records []recordStruct, err error) {
	response, err := f.client.DescribeDomainRecords(request)
	if err != nil {
		return
	}

	for _, record := range response.DomainRecords.Record {
		tempRecords = append(tempRecords, recordStruct{rr: record.RR, domain: record.DomainName})
	}

	if int(response.TotalCount) > len(tempRecords) {
		value, err := request.PageNumber.GetValue()

		if err != nil {
			return []recordStruct{}, err
		}

		request.PageNumber = requests.NewInteger(value + 1)

		return f.fetchDomainRecords(request, tempRecords)
	}

	return tempRecords, nil
}
