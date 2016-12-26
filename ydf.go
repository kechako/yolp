package yolp

import (
	"fmt"
	"time"
)

// YDFファイルのルートノード。
type YDF struct {
	ResultInfo Result
	Feature    []Feature
	Dictionary Dictionary
}

// YDF データのレスポンス情報全体に関する情報を表す要素。
type Result struct {
	Count       uint64
	Total       uint64
	Start       uint64
	Latency     float32
	Status      uint64
	Description string
	Copyright   string
}

// 地図上に表示されるアイコンなどの図形の情報。
type Feature struct {
	Id          string
	Name        string
	Category    []string
	Description string
	Geometry    Geometry
	Property    Property
	Style       Style
	RouteInfo   []Route
}

// 図形の種類や緯度経度など地理的な情報。
type Geometry struct {
	Id           string
	Target       string
	Type         GeometryType
	Coordinates  string
	BoundingBox  string
	Compress     string
	CompressType string
	Datum        Datum
	Exterior     Polygon
	Interior     Polygon
	Radius       string
	Geometry     []Geometry
}

// Geometryのタイプ。
type GeometryType string

const (
	Point         GeometryType = "point"
	LineString    GeometryType = "linestring"
	PolygonType   GeometryType = "polygon"
	Circle        GeometryType = "circle"
	Ellipse       GeometryType = "ellipse"
	MultiGeometry GeometryType = "multigeometry"
)

// 測地系。
type Datum string

const (
	WGS Datum = "wgs"
	TKY Datum = "tkw"
)

// ポリゴン
type Polygon struct {
	Coordinates string
}

// 地域・拠点情報の詳細。
type Property struct {
	WeatherAreaCode int
	WeatherList     WeatherList
}

type WeatherList struct {
	Weather []Weather
}

type Weather struct {
	Type     WeatherType
	Date     string
	Rainfall float32
}

type WeatherType string

const (
	Observation WeatherType = "observation"
	Forecast    WeatherType = "forecast"
)

type Route struct {
	Edge     []Edge
	Property Property
}

type Edge struct {
	Id       string
	Vertex   []Vertex
	Property Property
}

type Vertex struct {
	Type     VertexType
	Property Property
}

type VertexType string

const (
	Start VertexType = "Start"
	End   VertexType = "End"
)

// Style要素のコンテナとなる要素
type Dictionary struct {
	Style []Style
}

// 地図上に表示されるアイコンなどの図形のスタイルの情報
type Style struct {
	Id        string
	Target    string
	Type      StyleType
	Image     string
	Size      string
	Anchor    string
	Opacity   float32
	Color     string
	StartLine LineEnd
	EndLine   LineEnd
}

// スタイルのタイプ。
type StyleType string

const (
	Icon StyleType = "icon"
	Line StyleType = "line"
	Fill StyleType = "fill"
)

type LineEnd string

const (
	Arrow LineEnd = "arrow"
)

func (w *Weather) IsObservation() bool {
	return w.Type == Observation
}

func (w *Weather) IsForecast() bool {
	return w.Type == Forecast
}

func (w *Weather) IsRaining() bool {
	return w.Rainfall > 0
}

func (w *Weather) Time() time.Time {
	var year, month, day, hour, min int

	_, err := fmt.Sscanf(w.Date, "%4d%2d%2d%2d%2d", &year, &month, &day, &hour, &min)
	if err != nil {
		return time.Time{}
	}

	return time.Date(year, time.Month(month), day, hour, min, 0, 0, time.Local)
}
