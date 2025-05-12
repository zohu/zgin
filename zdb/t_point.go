package zdb

import (
	"context"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"github.com/zohu/zgin/zutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type PointType string

const (
	PointTypeWGS84 PointType = "wgs84"
	PointTypeGCJ02 PointType = "gcj02"
	PointTypeBD09  PointType = "bd09"
)

// Point
// @Description: 数据库只存储WGS84
type Point struct {
	ewkb.Point
	Longitude float64   `json:"longitude,omitempty"`
	Latitude  float64   `json:"latitude,omitempty"`
	PointType PointType `json:"point_type"` // 仅仅用于描述Longitude和Latitude的类型，ewkb.Point一直都是WGS84
}

func NewPointFromWGS84(x, y float64) *Point {
	return &Point{
		Point: ewkb.Point{
			Point: geom.NewPointFlat(geom.XY, []float64{x, y}),
		},
		Longitude: x,
		Latitude:  y,
		PointType: PointTypeWGS84,
	}
}
func NewPointFromGCJ02(x, y float64) *Point {
	return NewPointFromWGS84(zutil.GCJ02toWGS84(x, y))
}
func NewPointFromDB09(x, y float64) *Point {
	return NewPointFromWGS84(zutil.BD09toWGS84(x, y))
}

func (p *Point) Equal(other Point) bool {
	return p.X() == other.X() && p.Y() == other.Y() && p.Z() == other.Z()
}

func (p *Point) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "GEOGRAPHY(Point)"
}

func (p *Point) Scan(value interface{}) error {
	if value == nil {
		*p = Point{}
		return nil
	}
	t, err := hex.DecodeString(value.(string))
	if err != nil {
		return err
	}
	err = p.Point.Scan(t)
	if err != nil {
		return err
	}
	p.Longitude = p.X()
	p.Latitude = p.Y()
	p.PointType = PointTypeWGS84
	return nil
}
func (p *Point) Value() (driver.Value, error) {
	return p.Point.Value()
}
func (p *Point) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL:  "ST_GeographyFromText(?)",
		Vars: []interface{}{fmt.Sprintf("POINT(%f %f)", p.X(), p.Y())},
	}
}

// GCJ02
// @Description: 高德、腾讯、阿里、google地图 坐标系
// @receiver p
// @return x
// @return y
func (p *Point) GCJ02() *Point {
	switch p.PointType {
	case PointTypeWGS84:
		p.Longitude, p.Latitude = zutil.WGS84toGCJ02(p.X(), p.Y())
	case PointTypeBD09:
		p.Longitude, p.Latitude = zutil.BD09toGCJ02(p.X(), p.Y())
	}
	p.PointType = PointTypeGCJ02
	return p
}

// BD09
// @Description: 百度坐标系
// @receiver p
// @return x
// @return y
func (p *Point) BD09() *Point {
	switch p.PointType {
	case PointTypeWGS84:
		p.Longitude, p.Latitude = zutil.WGS84toBD09(p.X(), p.Y())
	case PointTypeGCJ02:
		p.Longitude, p.Latitude = zutil.GCJ02toBD09(p.X(), p.Y())
	}
	p.PointType = PointTypeBD09
	return p
}

// WGS84
// @Description: WGS84坐标系
// @receiver p
// @return *Point
func (p *Point) WGS84() *Point {
	switch p.PointType {
	case PointTypeGCJ02:
		p.Longitude, p.Latitude = zutil.GCJ02toWGS84(p.X(), p.Y())
	case PointTypeBD09:
		p.Longitude, p.Latitude = zutil.BD09toWGS84(p.X(), p.Y())
	}
	p.PointType = PointTypeWGS84
	return p
}
