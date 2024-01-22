package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/corona10/goimagehash"
	"github.com/kbinani/screenshot"
)

// 保存图片到文件
func SaveImage(img image.Image, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	png.Encode(f, img)
}

// 获取所有屏幕的截图
func GetScreenShots() ([]image.Image, error) {
	n := screenshot.NumActiveDisplays()

	screenshots := make([]image.Image, n)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			return nil, err
		}
		screenshots[i] = img
	}
	return screenshots, nil
}

// 将image.Image转换为*image.RGBA
func ToRGBA(img image.Image) *image.RGBA {
	if dst, ok := img.(*image.RGBA); ok {
		return dst
	}

	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	return dst
}

// 计算两张图片的相似度
func MatchImage(target *image.Image, pattern image.Image) float64 {
	hash1, err := goimagehash.PerceptionHash(*target)
	if err != nil {
		return -1
	}

	hash2, err := goimagehash.PerceptionHash(pattern)
	if err != nil {
		return -1
	}

	distance, err := hash1.Distance(hash2)
	if err != nil {
		return -1
	}

	return 1 - float64(distance)/64
}

// 从文件中读取图片
func LoadImage(filename string) image.Image {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	return img
}

// 从base64字符串中读取图片
func LoadBase64Image(base64Str string) image.Image {
	// Decode the base64 string
	unbased, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		log.Fatal("error decoding base64 string: %w", err)
	}

	// Convert the bytes to an io.Reader
	r := strings.NewReader(string(unbased))

	// Decode the image
	img, _, err := image.Decode(r)
	if err != nil {
		log.Fatal("error decoding image: %w", err)
	}

	return img
}

type Crop struct {
	leftRatio   float32
	rightRatio  float32
	topRatio    float32
	aspectRatio float32
	fromTop     bool
}

// 获取裁剪后的矩形
func (crop Crop) getCropRect(window Recti32) image.Rectangle {
	x1 := float32(window.x1)
	y1 := float32(window.y1)
	x2 := float32(window.x2)
	y2 := float32(window.y2)

	width := x2 - x1

	nx1 := x1 + width*crop.leftRatio
	nx2 := x2 - width*crop.rightRatio

	var ny1, ny2 float32

	if crop.fromTop {
		ny1 = y1 + width*crop.topRatio
		ny2 = ny1 + (nx2-nx1)/crop.aspectRatio
	} else {
		ny2 = y2 - width*crop.topRatio
		ny1 = ny2 - (nx2-nx1)/crop.aspectRatio
	}

	cropRect := image.Rect(int(nx1), int(ny1), int(nx2), int(ny2))
	return cropRect
}

// 裁剪图片，仅适用于白荆回廊
func (crop Crop) CutImage(img *image.Image, window Recti32) image.Image {
	cropRect := crop.getCropRect(window)
	croppedImg := (*img).(*image.RGBA).SubImage(cropRect).(*image.RGBA)
	return croppedImg
}

// 获取中心位置
func (crop Crop) GetCenter(window Recti32) (float32, float32) {
	cropRect := crop.getCropRect(window)
	return float32(cropRect.Min.X+cropRect.Max.X) / 2, float32(cropRect.Min.Y+cropRect.Max.Y) / 2
}

// 用于美化Map的输出
func BeautifulMap(m map[string]float64) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%s: %.2f", k, m[k]))
	}

	return sb.String()
}
