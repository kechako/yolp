package yolp

import (
	"fmt"
	"image"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/xerrors"
)

const yolpStaticUrl = "https://map.yahooapis.jp/map/V1/static"

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

		if options.Mode != MapModeNormal {
			query["mode"] = url.QueryEscape(string(options.Mode))
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
		for _, pin := range options.Pins {
			query[pin.QueryKey()] = pin.String()
		}
		if options.Overlay != nil {
			query["overlay"] = options.Overlay.String()
		}
	}

	return y.apiGetImage(y.makeUrl(yolpStaticUrl, query))
}

type MapMode string

const (
	MapModeNormal      MapMode = ""
	MapModePhoto               = "photo"
	MapModeUnderground         = "map-b1"
	MapModeHD                  = "hd"
	MapModeHybrid              = "hybrid"
	MapModeBlank               = "blankmap"
	MapModeOSM                 = "osm"
)

type StaticOptions struct {
	Mode    MapMode
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
	if options.Zoom != 0 {
		zoomMin, zoomMax := 1, 20
		switch options.Mode {
		case MapModeUnderground:
			zoomMin, zoomMax = 19, 21
		case MapModeBlank:
			zoomMin, zoomMax = 11, 20
		}
		if options.Zoom < zoomMin || options.Zoom > zoomMax {
			return xerrors.Errorf("Zoom must be between %d to %d if it isn't 0", zoomMin, zoomMax)
		}
	}
	if len(options.Pins) > 0 {
		for _, pin := range options.Pins {
			if err := pin.IsValid(); err != nil {
				return err
			}
		}
	}
	if options.Overlay != nil {
		if options.Overlay.Type == OverlayTypeRainfall && options.Zoom > 15 {
			return xerrors.New("Zoom must be between 1 to 20 if overlay rainfall")
		}
	}

	return nil
}

type PinStyle int

const (
	PinStyleNormal PinStyle = iota
	PinStyleNumber
	PinStyleAlphabet
	PinStyleStar
)

type PinColor string

const (
	PinColorDefault PinColor = ""
	PinColorRed              = "red"
	PinColorBlue             = "blue"
	PinColorGreen            = "green"
	PinColorYellow           = "yellow"
)

type Pin struct {
	Style     PinStyle
	Latitude  float32
	Longitude float32
	Label     string
	Color     PinColor
	Number    int
	Alphabet  rune
}

func (p *Pin) IsValid() error {
	switch p.Style {
	case PinStyleNumber:
		if p.Number < 0 || p.Number > 99 {
			return xerrors.New("Pin.Number must be between 0 to 99 if the style is number.")
		}
	case PinStyleAlphabet:
		if p.Alphabet < 'a' || p.Alphabet > 'z' {
			return xerrors.New("Pin.Alphabet must be between 'a' to 'z' if the style is alphabet.")
		}
	}

	return nil
}

func (p *Pin) QueryKey() string {
	switch p.Style {
	case PinStyleNormal:
		return "pin"
	case PinStyleNumber:
		if p.Number < 0 || p.Number > 99 {
			return "pin"
		}

		return fmt.Sprintf("pin%d", p.Number)
	case PinStyleAlphabet:
		if p.Alphabet < 'a' || p.Alphabet > 'z' {
			return "pin"
		}

		return fmt.Sprintf("pin%c", p.Alphabet)
	case PinStyleStar:
		return "pindefault"
	}

	return "pin"
}

func (p *Pin) String() string {
	var values []string
	values = append(values, fmt.Sprintf("%f", p.Latitude))
	values = append(values, fmt.Sprintf("%f", p.Longitude))
	if p.Label != "" {
		values = append(values, url.QueryEscape(p.Label))
	}
	if p.Color != PinColorDefault {
		values = append(values, url.QueryEscape(string(p.Color)))
	}

	return strings.Join(values, ",")
}

type OverlayType string

const (
	OverlayTypeNone     OverlayType = ""
	OverlayTypeRainfall OverlayType = "rainfall"
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
