package yolp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type YOLP struct {
	appID  string
	client *http.Client
}

func NewYOLP(appID string) *YOLP {
	return &YOLP{
		appID:  appID,
		client: &http.Client{},
	}
}

func NewYOLPWithClient(appID string, client *http.Client) *YOLP {
	return &YOLP{
		appID:  appID,
		client: client,
	}
}

func (y *YOLP) AppID() string {
	return y.appID
}

const (
	yolpPlaceUrl         = "http://weather.olp.yahooapis.jp/v1/place"
	yolpSearchZipCodeUrl = "http://search.olp.yahooapis.jp/OpenLocalPlatform/V1/zipCodeSearch"
)

func (y *YOLP) Place(latitude float32, longitude float32) (*YDF, error) {
	query := map[string]string{
		"coordinates": fmt.Sprintf("%f,%f", longitude, latitude),
		"interval":    "5",
	}

	return y.apiGet(y.makeUrl(yolpPlaceUrl, query))
}

func (y *YOLP) SearchZipCode(zipCode string) (*YDF, error) {
	query := map[string]string{
		"query": zipCode,
	}

	return y.apiGet(y.makeUrl(yolpSearchZipCodeUrl, query))
}

func (y *YOLP) apiGet(url string) (*YDF, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Can not create a new request")
	}

	res, err := y.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	ydf := &YDF{}
	err = d.Decode(ydf)
	if err != nil {
		return nil, err
	}

	return ydf, nil
}

func (y *YOLP) makeUrl(baseUrl string, query map[string]string) string {
	u, err := url.Parse(baseUrl)
	if err != nil {
		log.Panic(err)
	}

	q := u.Query()

	for key, value := range query {
		q.Add(key, value)
	}
	q.Add("output", "json")
	q.Add("appid", y.appID)
	u.RawQuery = q.Encode()

	return u.String()
}
