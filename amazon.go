package amazonProductAdvertisingAPI

// Amazon Product Advertising API
// http://aws.amazon.com/archives/Product-Advertising-API/8967000559514506
// http://docs.aws.amazon.com/AWSECommerceService/latest/DG/ItemSearch.html

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type Connection struct {
	AccessKeyID     string
	AccessKeySecret string
	AssociateTag    string
	Region          string
}

type Image struct {
	XMLName xml.Name `xml:"MediumImage"`
	URL     string
	Height  uint16
	Width   uint16
}

type Item struct {
	XMLName       xml.Name `xml:"Item"`
	ASIN          string
	URL           string
	DetailPageURL string
	Author        string `xml:"ItemAttributes>Author"`
	Price         string `xml:"ItemAttributes>ListPrice>FormattedPrice"`
	PriceRaw      string `xml:"ItemAttributes>ListPrice>Amount"`
	Title         string `xml:"ItemAttributes>Title"`
	MediumImage   Image
}

type Request struct {
	XMLName           xml.Name          `xml:"Request"`
	IsValid           bool              `xml:"IsValid"`
	ItemLookupRequest ItemLookupRequest `xml:"ItemLookupRequest"`
}

type ItemLookupRequest struct {
	XMLName xml.Name `xml:"ItemLookupRequest"`
}

type ItemLookupResponse struct {
	XMLName xml.Name `xml:"ItemLookupResponse"`
	Items   []Item   `xml:"Items>Item"`
	Request Request  `xml:"Items>Request"`
}

type ItemSearchResponse struct {
	XMLName xml.Name `xml:"ItemSearchResponse"`
	Items   []Item   `xml:"Items>Item"`
	Request Request  `xml:"Items>Request"`
}

var service_domains = map[string]string{
	"CA": "ecs.amazonaws.ca",
	"CN": "webservices.amazon.cn",
	"DE": "ecs.amazonaws.de",
	"ES": "webservices.amazon.es",
	"FR": "ecs.amazonaws.fr",
	"IT": "webservices.amazon.it",
	"JP": "ecs.amazonaws.jp",
	"UK": "ecs.amazonaws.co.uk",
	"US": "ecs.amazonaws.com",
}

func (self Connection) requestForArguments(arguments map[string]string) (req http.Request, err error) {
	arguments["AWSAccessKeyId"] = self.AccessKeyID
	arguments["AssociateTag"] = self.AssociateTag
	arguments["Version"] = "2011-08-01"
	arguments["Service"] = "AWSEcommerceService"
	arguments["Timestamp"] = time.Now().Format(time.RFC3339)
	keys := make([]string, 0, len(arguments))
	for argument := range arguments {
		keys = append(keys, argument)
	}
	sort.Strings(keys)
	// build the query string
	var queryString string
	for i, key := range keys {
		if i > 0 {
			queryString += "&"
		}
		queryString += key + "=" + url.QueryEscape(arguments[key])
	}
	// hash and sign the query string
	domain := service_domains[self.Region]
	data := "GET\n" + domain + "\n/onca/xml\n" + queryString
	hash := hmac.New(sha256.New, []byte(self.AccessKeySecret))
	hash.Write([]byte(data))
	signature := url.QueryEscape(base64.StdEncoding.EncodeToString(hash.Sum(nil)))
	queryString += "&Signature=" + signature
	// create and return the request object
	requestURL := fmt.Sprintf("http://%s/onca/xml?%s", domain, queryString)
	newreq, err := http.NewRequest("GET", requestURL, nil)
	req = *newreq
	return req, err
}

func (self Connection) performRequest(req *http.Request) (contents []byte, err error) {
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return
	}
	contents, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	defer response.Body.Close()

	fmt.Printf("status code %v\n", response.StatusCode)
	fmt.Println(string(contents))

	return contents, err
}

func (self Connection) ItemLookup(arguments map[string]string) (itemLookupResponse ItemLookupResponse, err error) {
	itemLookupResponse = ItemLookupResponse{}
	req, err := self.requestForArguments(arguments)
	if err != nil {
		return
	}
	responseData, err := self.performRequest(&req)
	if err != nil {
		return
	}
	err = xml.Unmarshal(responseData, &itemLookupResponse)
	return
}

func (self Connection) ItemSearch(arguments map[string]string) (itemSearchResponse ItemSearchResponse, err error) {
	itemSearchResponse = ItemSearchResponse{}
	req, err := self.requestForArguments(arguments)
	if err != nil {
		return
	}
	responseData, err := self.performRequest(&req)
	if err != nil {
		return
	}
	err = xml.Unmarshal(responseData, &itemSearchResponse)
	return
}
