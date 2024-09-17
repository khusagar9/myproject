package noflyzone

import "time"

// Helper functions for geometry
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Function to check if two line segments intersect
func doLinesIntersect(l1, l2 Line) bool {
	// Get the orientation of the triplet (p, q, r)
	orientation := func(p, q, r Point) int {
		val := (q.Lon-p.Lon)*(r.Lat-p.Lat) - (q.Lat-p.Lat)*(r.Lon-p.Lon)
		if val == 0 {
			return 0 // collinear
		}
		if val > 0 {
			return 1 // clockwise
		}
		return 2 // counterclockwise
	}

	// Check if point q lies on segment pr
	onSegment := func(p, q, r Point) bool {
		return q.Lon <= max(p.Lon, r.Lon) && q.Lon >= min(p.Lon, r.Lon) &&
			q.Lat <= max(p.Lat, r.Lat) && q.Lat >= min(p.Lat, r.Lat)
	}

	// Get orientations of the four triplets
	o1 := orientation(l1.Start, l1.End, l2.Start)
	o2 := orientation(l1.Start, l1.End, l2.End)
	o3 := orientation(l2.Start, l2.End, l1.Start)
	o4 := orientation(l2.Start, l2.End, l1.End)

	// General case: the lines intersect
	if o1 != o2 && o3 != o4 {
		return true
	}

	// Special cases: collinear points
	if o1 == 0 && onSegment(l1.Start, l2.Start, l1.End) {
		return true
	}
	if o2 == 0 && onSegment(l1.Start, l2.End, l1.End) {
		return true
	}
	if o3 == 0 && onSegment(l2.Start, l1.Start, l2.End) {
		return true
	}
	if o4 == 0 && onSegment(l2.Start, l1.End, l2.End) {
		return true
	}

	// If none of the cases matched, the lines do not intersect
	return false
}

// Function to check if a path intersects a polygon
func doesPathIntersectPolygon(path Line, polygon []Point) bool {
	for i := 0; i < len(polygon); i++ {
		// Get the next point (close the loop)
		next := (i + 1) % len(polygon)
		// Create a line for the current edge of the polygon
		polygonEdge := Line{Start: polygon[i], End: polygon[next]}
		// Check if the path intersects with the current edge
		if doLinesIntersect(path, polygonEdge) {
			return true
		}
	}
	return false
}

// IsPathInNoFlyZone checks if the path intersects any active no-fly zone.
func IsPathInNoFlyZone(source, destination Point, noFlyZones []NoFlyZone, currentTime time.Time) bool {
	// Define the path as a line segment
	path := Line{Start: source, End: destination}

	// Loop through all no-fly zones
	for _, zone := range noFlyZones {
		// Check if the no-fly zone is active
		if currentTime.After(zone.StartTime) && currentTime.Before(zone.EndTime) {
			// Check if the path intersects the no-fly zone polygon
			if doesPathIntersectPolygon(path, zone.Polygon) {
				return true
			}
		}
	}
	return false
}
