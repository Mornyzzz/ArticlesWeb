package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	// Create a new image with a size of 300x300 pixels
	img := image.NewRGBA(image.Rect(0, 0, 300, 300))

	p1 := image.Point{150, 50}
	p2 := image.Point{50, 250}
	p3 := image.Point{250, 250}

	q1 := image.Point{150, 250}
	q2 := image.Point{50, 50}
	q3 := image.Point{250, 50}

	for x := 0; x < 300; x++ {
		for y := 0; y < 300; y++ {
			p := image.Point{x, y}
			if isInsideTriangle(p, p1, p2, p3) {
				img.Set(x, y, color.RGBA{94, 94, 94, 255}) // Red color
			}
			if isInsideTriangle(p, q1, q2, q3) {
				img.Set(x, y, color.RGBA{0, 156, 165, 255}) // Синий цвет
			}
		}
	}

	file, _ := os.Create("amazing_logo.png")
	defer file.Close()

	png.Encode(file, img)
}

// Helper function to determine if a point is inside the triangle
func isInsideTriangle(p, a, b, c image.Point) bool {
	return crossProduct(p, a, b)*crossProduct(p, b, c) >= 0 &&
		crossProduct(p, b, c)*crossProduct(p, c, a) >= 0 &&
		crossProduct(p, c, a)*crossProduct(p, a, b) >= 0
}

// Helper function to calculate the cross product of two vectors
func crossProduct(p, q, r image.Point) int {
	return (q.X-p.X)*(r.Y-p.Y) - (q.Y-p.Y)*(r.X-p.X)
}
