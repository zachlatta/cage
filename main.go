package main

import (
  "flag"
  "fmt"
  "image"
  _ "image/png"
  "os"
  "io"
  "log"
)

type Region struct {
  Bounds image.Rectangle
  AverageLuminance float64
  X int
  Y int
}


func main() {
  flag.Usage = usage
  flag.Parse()
  log.SetFlags(0)

  args := flag.Args()
  if len(args) < 1 {
    usage()
  }

  fImg, err := os.Open(args[0]);
  if err != nil {
    imageError()
  }
  defer fImg.Close()

  img, _, err := image.Decode(fImg)
  if err != nil {
    imageError()
  }

  characters := map[float64]string {
    5.2: " ",
    8.0: "#",
    8.5: "@",
    10.0: "O",
    100.0: " ",
  }

  width := 25 

  regionSize := img.Bounds().Dx() / width
  regions := make([]Region, regionCount(img, regionSize))

  i := 0
  eachRegion(img, regionSize, func(x, y int, region image.Image) {
    regions[i] = Region{region.Bounds(), 0.0, x, y}

    sumLuminance, count := 0.0, 0.0
    eachPixel(region, func(pX, pY int, r, g, b, a uint8) {
      fR, fG, fB := float64(r), float64(g), float64(b)
      sumLuminance += (fR + fG + fB) / 3
      count++
    })
    regions[i].AverageLuminance = sumLuminance / count

    i++
  })

  output := make([][]string, img.Bounds().Dx() / regionSize)
  for i := range(output) {
    output[i] = make([]string, img.Bounds().Dy() / regionSize)
  }

  for _, r := range(regions) {
    for k, v := range(characters) {
      if r.AverageLuminance <= k * float64(width) {
        output[r.X][r.Y] = v
        break
      }
    }
  }

  for i := 0; i < len(output[0]); i++ {
    for _, x := range(output) {
      fmt.Print(x[i])
    }
    fmt.Println()
  }
}

func eachPixel(img image.Image, fn func(x, y int, r, g, b, a uint8)) {
  rect := img.Bounds()
  for i := 0; i < rect.Dx() * rect.Dy(); i++ {
    x, y := i % rect.Dx() + rect.Min.X, i / rect.Dx() + rect.Min.Y
    pixel := img.At(x, y)
    r32, g32, b32, a32 := pixel.RGBA()
    r8, g8, b8, a8 := uint8(r32), uint8(g32), uint8(b32), uint8(a32)
    fn(x, y, r8, g8, b8, a8)
  }
}

func eachRegion(img image.Image, regionSize int, fn func(x, y int, region image.Image)) {
  rect := img.Bounds()
  regionsX, regionsY := rect.Dx() / regionSize, rect.Dy() / regionSize
  for i := 0; i < regionCount(img, regionSize); i++ {
    x, y := i % regionsX + rect.Min.X, i / regionsY + rect.Min.Y
    regionX, regionY := x * regionSize, y * regionSize
    rect := image.Rect(regionX, regionY,
                       regionX + regionSize, regionY + regionSize)
    region := img.(interface {
      SubImage(r image.Rectangle) image.Image
    }).SubImage(rect)
    fn(x, y, region)
  }
}

func regionCount(img image.Image, regionSize int) int {
  rect := img.Bounds()
  return (rect.Dx() / regionSize) * (rect.Dy() / regionSize)
}

var usageText = `cage creates ascii art from images.

Usage:

  cage [image name] [flags]

The available flags are:

  -w  width of output in characters, default is 50
`

func printUsage(w io.Writer) {
  fmt.Fprintln(w, usageText)
}

func usage() {
  printUsage(os.Stderr)
  os.Exit(2)
}

func imageError() {
  fmt.Fprintln(os.Stderr, "Error opening image.")
  os.Exit(2)
}
