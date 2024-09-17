package utils

import (
	"myproject/models"
)

// IsPathIntersectingZone checks if a straight path between source and destination intersects a polygon zone
func IsPathIntersectingZone(source, dest models.Coordinate, zone models.NoFlyZone) bool {
	for _, polygon := range zone.Coordinates {
		for i := 0; i < len(polygon[0])-1; i++ {
			if isLineIntersecting(source, dest, polygon[0][i], polygon[0][i+1]) {
				return true
			}
		}
	}
	return false
}

// isLineIntersecting checks if two line segments (p1-p2 and q1-q2) intersect
func isLineIntersecting(p1, p2, q1, q2 models.Coordinate) bool {
	// Check the orientation of the points
	o1 := orientation(p1, p2, q1)
	o2 := orientation(p1, p2, q2)
	o3 := orientation(q1, q2, p1)
	o4 := orientation(q1, q2, p2)

	// General case: if orientations are different, the lines intersect
	if o1 != o2 && o3 != o4 {
		return true
	}

	// Special cases: check for collinear points
	if o1 == 0 && onSegment(p1, q1, p2) {
		return true
	}
	if o2 == 0 && onSegment(p1, q2, p2) {
		return true
	}
	if o3 == 0 && onSegment(q1, p1, q2) {
		return true
	}
	if o4 == 0 && onSegment(q1, p2, q2) {
		return true
	}

	return false
}

// orientation finds the orientation of the ordered triplet (p, q, r)
// Returns:
// 0 -> p, q and r are collinear
// 1 -> Clockwise
// 2 -> Counterclockwise
func orientation(p, q, r models.Coordinate) int {
	val := (q.Lat-p.Lat)*(r.Lon-q.Lon) - (q.Lon-p.Lon)*(r.Lat-q.Lat)
	if val == 0 {
		return 0 // Collinear
	} else if val > 0 {
		return 1 // Clockwise
	} else {
		return 2 // Counterclockwise
	}
}

// onSegment checks if point q lies on line segment pr
func onSegment(p, q, r models.Coordinate) bool {
	if q.Lon <= max(p.Lon, r.Lon) && q.Lon >= min(p.Lon, r.Lon) &&
		q.Lat <= max(p.Lat, r.Lat) && q.Lat >= min(p.Lat, r.Lat) {
		return true
	}
	return false
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
