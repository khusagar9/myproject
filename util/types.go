package util

// Point represents a latitude/longitude coordinate.
type Point struct {
	Lat, Lon float64
}

// Line represents a line segment between two points.
type Line struct {
	Start, End Point
}

// RestrictedZone represents a polygon with a time range when it is active.
type RestrictedZone struct {
	ID        string
	Polygon   []Point
	StartTime int64
	EndTime   int64
}
