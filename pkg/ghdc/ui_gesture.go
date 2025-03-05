package ghdc

import (
	"math"
)

var EVENT_INJECTION_DELAY_MS = 50

type Point struct {
	X int `json:"x,omitempty"`
	Y int `json:"y,omitempty"`
}

func (p Point) distance(p0 Point) float64 {
	return math.Sqrt(math.Pow(float64(p.X-p0.X), 2) + math.Pow(float64(p.Y-p0.Y), 2))
}

type GestureStep struct {
	Point
	GestureType string
	Duration    int
}

type EmptyGesture struct{}

type Gesture struct {
	steps []*GestureStep
}

type PointerMatrix struct {
	points []*FingerPoint
}

type FingerPoint struct {
	index int
	point Point
}

func NewGesture() *EmptyGesture {
	return new(EmptyGesture)
}

func (g *EmptyGesture) Start(point Point) *Gesture {
	gesture := &Gesture{}
	gesture.steps = append(gesture.steps, &GestureStep{Point: point, GestureType: "start"})
	return gesture
}

func (g *Gesture) MoveTo(point Point, duration int) *Gesture {
	g.steps = append(g.steps, &GestureStep{Point: point, GestureType: "move", Duration: duration})
	return g
}

func (g *Gesture) Pause(duration int) *Gesture {
	g.steps = append(g.steps, &GestureStep{GestureType: "pause", Duration: duration})
	return g
}

func (gs *GestureStep) calculateSteps() int {
	return (gs.Duration / EVENT_INJECTION_DELAY_MS) + 1
}

func (pm *PointerMatrix) setPoint(point Point, fingerIndex, duration int) {
	point.X += 65536 * duration
	pm.points = append(pm.points, &FingerPoint{index: fingerIndex, point: point})
}

func (pm *PointerMatrix) fingerIndexStats() (fingers int, maxSteps int) {
	indexMap := make(map[int]int)

	for _, fp := range pm.points {
		indexMap[fp.index]++
	}

	fingers = len(indexMap)
	for _, count := range indexMap {
		if count > maxSteps {
			maxSteps = count
		}
	}

	return fingers, maxSteps
}
