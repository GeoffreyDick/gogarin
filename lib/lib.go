package lib

import (
	"errors"
	"log"
	"math"

	m "github.com/GeoffreyDick/gogarin/model"
)

// Filter filters a slice of any type by a given function.
func Filter[T any](slice []T, f func(T) bool) []T {
	var n []T
	for _, e := range slice {
		if f(e) {
			n = append(n, e)
		}
	}
	return n
}

// Contains checks if a string is in a slice of strings.
func Contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

type Coordinate struct {
	x int
	y int
}

// Distance calculates the distance between two xy coordinates.
func Distance(c1, c2 Coordinate) float64 {
	return math.Sqrt(math.Pow(float64(c1.x-c2.x), 2) + math.Pow(float64(c1.y-c2.y), 2))
}

// NearestWaypoint returns the nearest waypoint to a given coordinate.
func NearestWaypoint(currentWaypoint *m.Waypoint, waypoints *[]m.Waypoint) (*m.Waypoint, error) {
	var nearestWaypoint *m.Waypoint
	var nearestDistance float64
	log.Printf("current waypoint: %v", currentWaypoint)
	log.Printf("waypoints: %v", waypoints)

	for _, waypoint := range *waypoints {
		distance := Distance(Coordinate{currentWaypoint.X, currentWaypoint.Y}, Coordinate{waypoint.X, waypoint.Y})
		if nearestWaypoint == nil || distance < nearestDistance {
			nearestWaypoint = &waypoint
			nearestDistance = distance
		}
	}

	if nearestWaypoint == nil {
		return nil, errors.New("no nearest waypoint found")
	}

	return nearestWaypoint, nil
}
