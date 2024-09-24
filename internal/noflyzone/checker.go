package noflyzone

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

// IsPathInNoFlyZone checks if the path intersects any active no-fly zone.
func IsPathInNoFlyZone(source, destination Point, noFlyZones []NoFlyZone) bool {
	// Define the path as a line segment
	path := Line{Start: source, End: destination}
	// Loop through all no-fly zones
	for _, zone := range noFlyZones {
		if doesLineIntersectPolygon(path, zone) {
			return true
		}
	}
	return false
}

func doesLineIntersectPolygon(line Line, noFlyZone NoFlyZone) bool {
	polygon := noFlyZone.Polygon
	n := len(polygon)

	// Check if the line segment intersects with any edge of the polygon
	for i := 0; i < n; i++ {
		nextIndex := (i + 1) % n
		if doIntersect(line.Start, line.End, polygon[i], polygon[nextIndex]) {
			return true
		}
	}

	// Check if the line segment's start or end points are inside the polygon
	if isPointInPolygon(line.Start, polygon) || isPointInPolygon(line.End, polygon) {
		return true
	}

	return false
}

func isPointInPolygon(p Point, polygon []Point) bool {
	n := len(polygon)
	if n < 3 {
		return false
	}

	x, y := p.Lat, p.Lon
	inside := false

	p1x, p1y := polygon[0].Lat, polygon[0].Lon
	for i := 1; i <= n; i++ {
		p2x, p2y := polygon[i%n].Lat, polygon[i%n].Lon
		if y > min(p1y, p2y) && y <= max(p1y, p2y) && x <= max(p1x, p2x) {
			if p1y != p2y {
				xinters := (y-p1y)*(p2x-p1x)/(p2y-p1y) + p1x
				if p1x == p2x || x <= xinters {
					inside = !inside
				}
			}
		}
		p1x, p1y = p2x, p2y
	}

	return inside
}

func onSegment(p, q, r Point) bool {
	return q.Lat <= max(p.Lat, r.Lat) && q.Lat >= min(p.Lat, r.Lat) &&
		q.Lon <= max(p.Lon, r.Lon) && q.Lon >= min(p.Lon, r.Lon)
}

func orientation(p, q, r Point) int {
	val := (q.Lon-p.Lon)*(r.Lat-p.Lat) - (q.Lat-p.Lat)*(r.Lon-p.Lon)
	if val == 0 {
		return 0
	} else if val > 0 {
		return 1
	} else {
		return 2
	}
}

func doIntersect(p1, q1, p2, q2 Point) bool {
	o1 := orientation(p1, q1, p2)
	o2 := orientation(p1, q1, q2)
	o3 := orientation(p2, q2, p1)
	o4 := orientation(p2, q2, q1)

	if o1 != o2 && o3 != o4 {
		return true
	}

	if o1 == 0 && onSegment(p1, p2, q1) {
		return true
	}

	if o2 == 0 && onSegment(p1, q2, q1) {
		return true
	}

	if o3 == 0 && onSegment(p2, p1, q2) {
		return true
	}

	if o4 == 0 && onSegment(p2, q1, q2) {
		return true
	}

	return false
}
