package util

import (
	"h3d-drone-emulator/models"
	"math"
	"time"
)

var crossedZones []models.ClearanceZone

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

func haversineDistance(p1, p2 Point) float64 {
	const R = 6371e3 // Earth radius in meters

	lat1 := p1.Lat * math.Pi / 180 // Convert latitude to radians
	lon1 := p1.Lon * math.Pi / 180
	lat2 := p2.Lat * math.Pi / 180
	lon2 := p2.Lon * math.Pi / 180

	dLat := lat2 - lat1
	dLon := lon2 - lon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c // Distance in meters
}

// lineIntersectingRestrictedZones checks if the path intersects any active no-fly zone.
func lineIntersectingRestrictedZones(source, destination Point, polygon []Point) (Point, Point, bool) {
	var entry, exit Point
	foundEntry, foundExit := false, false

	for i := 0; i < len(polygon); i++ {
		next := (i + 1) % len(polygon)
		if doIntersect(source, destination, polygon[i], polygon[next]) {
			if !foundEntry {
				entry = fetchIntersectionPoint(source, destination, polygon[i], polygon[next])
				foundEntry = true
			} else if !foundExit {
				exit = fetchIntersectionPoint(source, destination, polygon[i], polygon[next])
				foundExit = true
			}
		}
	}

	// Return false if both entry and exit points are the same
	if foundEntry && foundExit && (entry != exit) {
		return entry, exit, true
	}
	return Point{}, Point{}, false
}

func fetchIntersectionPoint(p1, p2, q1, q2 Point) Point {
	// Implement the line intersection formula to calculate the intersection point
	// Assuming the lines are not parallel
	denom := (p1.Lat-p2.Lat)*(q1.Lon-q2.Lon) - (p1.Lon-p2.Lon)*(q1.Lat-q2.Lat)
	if denom == 0 {
		return Point{}
	}
	x := ((p1.Lat*p2.Lon-p1.Lon*p2.Lat)*(q1.Lon-q2.Lon) - (p1.Lon-p2.Lon)*(q1.Lat*q2.Lon-q1.Lon*q2.Lat)) / denom
	y := ((p1.Lat*p2.Lon-p1.Lon*p2.Lat)*(q1.Lat-q2.Lat) - (p1.Lat-p2.Lat)*(q1.Lat*q2.Lon-q1.Lon*q2.Lat)) / denom
	return Point{Lat: y, Lon: x}
}

func IsPathInRestrictedZone(source, destination Point, restrictedZones []RestrictedZone, currentTime time.Time, droneSpeedMilesPerHour float64) (bool, []models.ClearanceZone) {
	// Define the path as a line segment
	var isPathInRestrictedZone = false
	droneSpeedMeterPerSecond := convertMphToMps(droneSpeedMilesPerHour)
	// Loop through all no-fly zones
	for _, zone := range restrictedZones {

		entry, exit, intersects := lineIntersectingRestrictedZones(source, destination, zone.Polygon)
		if intersects {
			// Calculate entry time and exit time based on distance and speed
			entryDist := haversineDistance(source, entry)
			exitDist := haversineDistance(source, exit)
			entryTime := currentTime.Add(time.Duration(entryDist/droneSpeedMeterPerSecond) * time.Second)
			exitTime := currentTime.Add(time.Duration(exitDist/droneSpeedMeterPerSecond) * time.Second)

			// Ensure the entry and exit times are distinct and represent actual entry and exit points
			if (entryTime.Before(time.Unix(zone.EndTime, 0)) && exitTime.After(time.Unix(zone.StartTime, 0))) || (zone.EndTime == 0 || zone.StartTime == 0) {
				year, month, day := entryTime.Date()
				hour, minute, second := entryTime.Clock()

				// Convert to time.Date using the extracted components
				exitYear, exitMonth, exitDay := exitTime.Date()
				exitHour, exitMinute, exitSecond := exitTime.Clock()
				crossedZones = append(crossedZones, models.ClearanceZone{
					ID:        zone.ID,
					EntryTime: time.Date(year, month, day, hour, minute, second, 0, time.Local),
					ExitTime:  time.Date(exitYear, exitMonth, exitDay, exitHour, exitMinute, exitSecond, 0, time.Local),
				})
				isPathInRestrictedZone = true
			}
		}
	}
	return isPathInRestrictedZone, crossedZones
}

// Function to convert miles per hour to meters per second
func convertMphToMps(mph float64) float64 {
	return mph * 0.44704
}

func doLineIntersectPolygon(line Line, restrictedZone RestrictedZone) bool {
	polygon := restrictedZone.Polygon
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

// 0 (collinear), 1 (clockwise), 2 (counterclockwise)
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

func doIntersect(source, destination, polygonLineStart, polygonLineEnd Point) bool {
	o1 := orientation(source, destination, polygonLineStart)
	o2 := orientation(source, destination, polygonLineEnd)
	o3 := orientation(polygonLineStart, polygonLineEnd, source)
	o4 := orientation(polygonLineStart, polygonLineEnd, destination)

	if o1 != o2 && o3 != o4 {
		return true
	}

	if o1 == 0 && onSegment(source, polygonLineStart, destination) {
		return true
	}

	if o2 == 0 && onSegment(source, polygonLineEnd, destination) {
		return true
	}

	if o3 == 0 && onSegment(polygonLineStart, source, polygonLineEnd) {
		return true
	}

	if o4 == 0 && onSegment(polygonLineStart, destination, polygonLineEnd) {
		return true
	}

	return false
}
