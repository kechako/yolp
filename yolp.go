package yolp

import (
	"encoding/json"
	"fmt"
	"image"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/xerrors"
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
	yolpPlaceUrl         = "https://map.yahooapis.jp/weather/V1/place"
	yolpSearchZipCodeUrl = "https://map.yahooapis.jp/search/zip/V1/zipCodeSearch"
	yolpStaticUrl        = "https://map.yahooapis.jp/map/V1/static"
)

func (y *YOLP) Place(latitude float32, longitude float32) (*YDF, error) {
	query := map[string]string{
		"coordinates": fmt.Sprintf("%f,%f", longitude, latitude),
		"interval":    "5",
		"output":      "json",
	}

	return y.apiGet(y.makeUrl(yolpPlaceUrl, query))
}

type StaticOptions struct {
	Width   int
	Height  int
	Pointer bool
	Zoom    int
	Pins    []*Pin
	Overlay *Overlay
}

func (options *StaticOptions) IsValid() error {
	if options.Width < 0 {
		return xerrors.New("Width must be greater than 0 if it isn't 0")
	}
	if options.Height < 0 {
		return xerrors.New("Height must be greater than 0 if it isn't 0")
	}
	if options.Zoom < 0 || options.Zoom > 20 {
		return xerrors.New("Zoom must be between 1 to 20 if it isn't 0")
	}
	if options.Overlay != nil {
		if options.Overlay.Type == OverlayTypeRainfall && options.Zoom > 15 {
			return xerrors.New("Zoom must be between 1 to 20 if overlay rainfall")
		}
	}

	return nil
}

type Pin struct {
	Latitude  float32
	Longitude float32
	Label     string
	Color     PinColor
}

func (p *Pin) String() string {
	var values []string
	values = append(values, fmt.Sprintf("%f", p.Latitude))
	values = append(values, fmt.Sprintf("%f", p.Longitude))
	if p.Label != "" {
		values = append(values, url.QueryEscape(p.Label))
	}
	if p.Color != PinColorNone {
		values = append(values, url.QueryEscape(string(p.Color)))
	}

	return strings.Join(values, ",")
}

type PinColor string

const (
	PinColorNone   PinColor = ""
	PinColorRed             = "red"
	PinColorBlue            = "blue"
	PinColorGreen           = "green"
	PinColorYellow          = "yellow"
)

type Overlay struct {
	Type      OverlayType
	Date      time.Time
	DateLabel bool
}

func (o *Overlay) String() string {
	var str strings.Builder

	oType := o.Type
	if oType == OverlayTypeNone {
		oType = OverlayTypeRainfall
	}

	str.WriteString("type:")
	str.WriteString(url.QueryEscape(string(oType)))

	if !o.Date.IsZero() {
		str.WriteString("|date:")
		str.WriteString(o.Date.Format("20060102"))
	}

	str.WriteString("|datelabel:")
	if o.DateLabel {
		str.WriteString("on")
	} else {
		str.WriteString("off")
	}

	return str.String()
}

type OverlayType string

const (
	OverlayTypeNone     OverlayType = ""
	OverlayTypeRainfall OverlayType = "rainfall"
)

// Static gets a static map image from Yahoo! Static Map API.
// API document : https://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/static.html
func (y *YOLP) Static(latitude, longitude float32, options *StaticOptions) (image.Image, error) {
	query := map[string]string{
		"lat": fmt.Sprintf("%f", latitude),
		"lon": fmt.Sprintf("%f", longitude),
	}

	if options != nil {
		if err := options.IsValid(); err != nil {
			return nil, err
		}

		if options.Width > 0 {
			query["width"] = strconv.Itoa(options.Width)
		}
		if options.Height > 0 {
			query["height"] = strconv.Itoa(options.Height)
		}
		if options.Pointer {
			query["pointer"] = "on"
		}
		if options.Zoom > 0 {
			query["z"] = strconv.Itoa(options.Zoom)
		}
		for i, pin := range options.Pins {
			key := fmt.Sprintf("pin%d", i+1)
			query[key] = pin.String()
		}
		if options.Overlay != nil {
			query["overlay"] = options.Overlay.String()
		}
	}

	return y.apiGetImage(y.makeUrl(yolpStaticUrl, query))
}

func (y *YOLP) SearchZipCode(zipCode string) (*YDF, error) {
	query := map[string]string{
		"query":  zipCode,
		"output": "json",
	}

	return y.apiGet(y.makeUrl(yolpSearchZipCodeUrl, query))
}

func (y *YOLP) apiGet(url string) (*YDF, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, xerrors.Errorf("can not create a new request: %w", err)
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

func (y *YOLP) apiGetImage(url string) (image.Image, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, xerrors.Errorf("can not create a new request: %w", err)
	}

	res, err := y.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	img, _, err := image.Decode(res.Body)
	if err != nil {
		return nil, err
	}

	return img, nil
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
	q.Add("appid", y.appID)
	u.RawQuery = q.Encode()

	return u.String()
}
