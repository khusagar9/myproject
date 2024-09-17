package noflyzone

import (
	"time"
)

// Point represents a latitude/longitude coordinate.
type Point struct {
	Lat, Lon float64
}

// Line represents a line segment between two points.
type Line struct {
	Start, End Point
}

// NoFlyZone represents a polygon with a time range when it is active.
type NoFlyZone struct {
	Polygon   []Point
	StartTime time.Time
	EndTime   time.Time
}
