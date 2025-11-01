package accessibility

import (
	"image"
	"testing"
)

func TestRectFromInfo(t *testing.T) {
	tests := []struct {
		name string
		info *ElementInfo
		want image.Rectangle
	}{
		{
			name: "Basic rectangle",
			info: &ElementInfo{
				Position: image.Point{X: 10, Y: 20},
				Size:     image.Point{X: 100, Y: 50},
			},
			want: image.Rect(10, 20, 110, 70),
		},
		{
			name: "Zero size",
			info: &ElementInfo{
				Position: image.Point{X: 5, Y: 5},
				Size:     image.Point{X: 0, Y: 0},
			},
			want: image.Rect(5, 5, 5, 5),
		},
		{
			name: "Negative position",
			info: &ElementInfo{
				Position: image.Point{X: -10, Y: -20},
				Size:     image.Point{X: 30, Y: 40},
			},
			want: image.Rect(-10, -20, 20, 20),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rectFromInfo(tt.info)
			if got != tt.want {
				t.Errorf("rectFromInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandRectangle(t *testing.T) {
	tests := []struct {
		name    string
		rect    image.Rectangle
		padding int
		want    image.Rectangle
	}{
		{
			name:    "Expand by 5",
			rect:    image.Rect(10, 10, 20, 20),
			padding: 5,
			want:    image.Rect(5, 5, 25, 25),
		},
		{
			name:    "Expand by 0",
			rect:    image.Rect(10, 10, 20, 20),
			padding: 0,
			want:    image.Rect(10, 10, 20, 20),
		},
		{
			name:    "Expand by negative",
			rect:    image.Rect(10, 10, 20, 20),
			padding: -2,
			want:    image.Rect(12, 12, 18, 18),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandRectangle(tt.rect, tt.padding)
			if got != tt.want {
				t.Errorf("expandRectangle() = %v, want %v", got, tt.want)
			}
		})
	}
}