//go:build darwin

package tray

import (
	"image/png"
	"os"
	"testing"
)

func TestDarwinTrayTemplateHasOpaqueAndTransparentPixels(t *testing.T) {
	file, err := os.Open("../../assets/tray_template.png")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	imageData, err := png.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	bounds := imageData.Bounds()
	if bounds.Dx() < 32 || bounds.Dy() < 32 {
		t.Fatalf("菜单栏图标尺寸过小: %v", bounds)
	}
	transparent := 0
	opaque := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := imageData.At(x, y).RGBA()
			switch alpha {
			case 0:
				transparent++
			case 0xffff:
				opaque++
			}
		}
	}
	if transparent == 0 || opaque == 0 {
		t.Fatalf("template 图标必须同时包含透明和不透明像素: transparent=%d opaque=%d", transparent, opaque)
	}
}
