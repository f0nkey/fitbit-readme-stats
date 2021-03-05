package main

import (
	"bytes"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgsvg"
	"image/color"
	"log"
	"strings"
	"time"
)

var (
	// Birds of Paradise https://tmtheme-editor.herokuapp.com/#!/editor/theme/Birds%20of%20Paradise
	ColorYellow = color.RGBA{R: 239, G: 172, B: 50, A: 255}
	ColorCream  = color.RGBA{R: 230, G: 225, B: 196, A: 255}
	ColorCoffee = color.RGBA{R: 50, G: 35, B: 35, A: 255}
	ColorOrange = color.RGBA{R: 239, G: 93, B: 50, A: 255}
)

// HTMLCode converts a color.RGBA to a format usable in CSS e.g., "rgba(5,25,50,1)"
func HTMLCode(c color.RGBA) string {
	// todo: normalize c.A to a valid CSS range [0,1]
	return fmt.Sprintf(`rgba(%d,%d,%d,%d)`, c.R, c.G, c.B, c.A)
}

// BannerXY represents the
type Banner struct {
	plot       *plot.Plot
	timeSeries plotter.XYs
	svg        string
}

// BannerXY represents a single point on the plot.
type BannerXY struct {
	X time.Time
	Y int
}

// BannerTicker is used to plot major and minor tick marks.
var BannerTicker = func(timeSeries plotter.XYs) plot.TickerFunc {
	return func(min, max float64) []plot.Tick {
		ticks := make([]plot.Tick, 0, len(timeSeries))
		for i, point := range timeSeries {
			if max-min <= 3600*2 { // do we have less than 2 hours of data?
				if i == 0 { // first tick, minute precision
					ticks = append(ticks, plot.Tick{
						Value: point.X,
						Label: "00:00",
					})
					continue
				}

				if int(point.X)%900 == 0 { // every 15 mins
					ticks = append(ticks, plot.Tick{
						Value: point.X,
						Label: "00:00",
					})
				}
				continue
			}

			if int(point.X)%3600 == 0 { // tick every hour
				ticks = append(ticks, plot.Tick{
					Value: point.X,
					Label: "00:00",
				})
			} else if int(point.X)%900 == 0 { // minor tick every 15 mins
				ticks = append(ticks, plot.Tick{
					Value: point.X,
					Label: "", // empty string == minor tick
				})
			}
		}
		return ticks
	}
}

// NewBanner returns a default banner displaying that the banner is not setup yet.
func NewBanner() *Banner {
	b := &Banner{}
	b.svg = defaultBanner(500, 100)
	return b
}

func defaultBanner(width, height int) string {
	bg := fmt.Sprintf(`<rect width="100%%" height="100%%" fill="%s" />`, HTMLCode(ColorCoffee))
	t := fmt.Sprintf(`<text x="%d" y="%d" fill="%s" style="font-family: sans-serif; font-weight:500;" dominant-baseline="hanging" text-anchor="middle">Banner not setup yet, or no data within range is available.</text>`, width/2, height/2, HTMLCode(ColorCream))
	banner := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" id="banner" width="%dpt" height="%dpt"> %s </svg>`, width, height, bg+t)
	return banner
}

// GenSVG generates SVG with data in arg xy.
// Use the SVG method to get the SVG image.
// Last x value in arg xy is used as current BPM in the middle of the heart.
func (b *Banner) GenSVG(xy []BannerXY, enableWatermark bool) (string, error) {
	timeSeries := make(plotter.XYs, 0, len(xy))
	for i := range xy {
		timeSeries = append(timeSeries, plotter.XY{
			X: float64(xy[i].X.Unix()),
			Y: float64(xy[i].Y),
		})
	}

	bpm := 0
	if len(timeSeries) <= 0 {
		b.svg = defaultBanner(500, 100)
		return b.svg, fmt.Errorf("data set empty")
	}
	bpm = int(timeSeries[len(timeSeries)-1].Y)

	// todo: use template/html instead of ugly sprintfs
	bannerWidth := 500 // in pts
	bannerHeight := 100
	thirdWidth := bannerWidth / 3 // heart takes up 1/3rd, plot 2/3rd
	topBottomPadding := 20        // pts of whitespace split between top and bottom

	plotWidth := thirdWidth * 2
	plotSVG := fmt.Sprintf(`<g transform="translate(%d,%d)"> %s </g>`, thirdWidth, 0, genPlot(timeSeries, plotWidth, bannerHeight))
	plotSVG = strings.ReplaceAll(plotSVG, `font-family:Times;font-weight:normal;font-style:normal;font-size:10px;`, "") // remove in-line style
	plotSVG = strings.ReplaceAll(plotSVG, `<?xml version="1.0"?>`, "")                                                  // cannot have multiple xml tags
	plotSVG = strings.ReplaceAll(plotSVG, "<text", `<text class="text"`)

	heartWidth := thirdWidth
	heartSVG := fmt.Sprintf(`<g transform="translate(%d %d)"> %s </g>`,
		0,
		-22,
		genHeart(bpm, heartWidth))
	heartTextSVG := fmt.Sprintf(`<g width="%d" height="%d" transform="translate(%d %d)"> %s </g>`,
		0,
		0,
		thirdWidth/2,
		bannerHeight/2,
		genHeartText(bpm, bannerHeight),
	)

	titleHeight := 12 // in pts
	titleSVG := genTitle("Heartrate From My FitBit Watch (Past 4 Hours)", titleHeight, bannerWidth/2)

	waterMarkSVG := ""
	if enableWatermark {
		waterMarkSVG = genWatermark("Get Source", 8, 5)
	}

	bg := fmt.Sprintf(`<rect width="100%%" height="100%%" fill="%s" />`, HTMLCode(ColorCoffee))
	style := fmt.Sprintf(`<style> .text {font: 600 9px "Arial", Sans-Serif; fill: %s;} </style>`, HTMLCode(ColorCream))
	banner := fmt.Sprintf(`<svg id="banner" xmlns="http://www.w3.org/2000/svg" width="%dpt" height="%dpt"> <!-- Generated via https://github.com/f0nkey/fitbit-readme-stats --> %s <g id="padding" transform="translate(0 %d)"> %s <g transform="translate(0 %d)"> %s </g> </g> </svg>`, bannerWidth, bannerHeight+titleHeight+topBottomPadding, bg+style, topBottomPadding/2, waterMarkSVG+titleSVG, titleHeight+6, plotSVG+heartSVG+heartTextSVG)
	b.svg = banner
	return b.svg, nil
}

// SVG returns the heart-rate banner in the svg format.
func (b *Banner) SVG() string {
	return b.svg
}

func genWatermark(text string, height, xOffset int) string {
	style := fmt.Sprintf(`font: 600 %dpt 'Arial', Sans-Serif; fill: %s;`, height, HTMLCode(ColorCream))
	return fmt.Sprintf(`<a href="https://github.com/f0nkey/fitbit-readme-stats"><text id="title" dominant-baseline="hanging" style="%s" x="%dpt"> %s </text></a>`, style, xOffset, text)
}

func genTitle(text string, height, xOffset int) string {
	style := fmt.Sprintf(`font: 600 %dpt 'Arial', Sans-Serif; fill: %s`, height, HTMLCode(ColorCream))
	return fmt.Sprintf(`<text id="title" dominant-baseline="hanging" text-anchor="middle" style="%s" x="%dpt"> %s </text>`, style, xOffset, text)
}

func genHeartText(bpm int, height int) string {
	bpmTextSize := 19 // points
	bpmTextOffset := height - bpmTextSize - 2
	text := fmt.Sprintf(`
		<text id="current-bpm-text" class="text" text-anchor="middle" x="%s" y="%d">Current BPM</text>
		<text id="bpm-number" class="text" dominant-baseline="middle" text-anchor="middle" x="%s" y="%s">%d</text>
		<style> #current-bpm-text {font-size: %dpt; fill: %s;}  #bpm-number {font-size: 35px; fill: %s;}</style>
`, "0", bpmTextOffset, "0", "0", bpm, bpmTextSize, HTMLCode(ColorCream), HTMLCode(ColorCoffee))

	return text
}

func genHeart(bpm int, width int) string {
	// https://codepen.io/tutsplus/pen/MLBMRw
	viewBox := width + width/3
	gOffset := viewBox / 2
	heart := fmt.Sprintf(`
	<svg width="%d" height="%d" viewBox="0 0 %d %d">
		<g transform="translate(%d %d)">
			<path transform="translate(-50 -50)" fill="%s" d="M92.71,7.27L92.71,7.27c-9.71-9.69-25.46-9.69-35.18,0L50,14.79l-7.54-7.52C32.75-2.42,17-2.42,7.29,7.27v0 c-9.71,9.69-9.71,25.41,0,35.1L50,85l42.71-42.63C102.43,32.68,102.43,16.96,92.71,7.27z"></path>
			<animateTransform 
			  attributeName="transform" 
			  type="scale" 
			  values="1; 1.5; 1.25; 1;" 
			  dur="%dms"
			  additive="sum"
			  repeatCount="indefinite">      
			</animateTransform>
		</g>
	</svg>
	`, width, width, viewBox, viewBox, gOffset, gOffset, HTMLCode(ColorYellow), 60000/bpm)

	return heart
}

func genPlot(timeSeries plotter.XYs, width, height int) string {
	p := plot.New()

	p.X.Tick.Marker = plot.TimeTicks{
		Ticker: BannerTicker(timeSeries),
		Format: "15:04",
		Time:   nil,
	}

	p.X.Tick.LineStyle.Color = ColorOrange
	p.X.LineStyle.Color = ColorOrange
	p.Y.Tick.LineStyle.Color = ColorOrange
	p.Y.LineStyle.Color = ColorOrange
	p.Y.Label.TextStyle.Color = ColorCream
	p.X.Label.TextStyle.Color = ColorCream
	p.X.Tick.Label.Color = ColorCream
	p.Y.Tick.Label.Color = ColorCream
	p.BackgroundColor = ColorCoffee

	line, err := plotter.NewLine(timeSeries)
	if err != nil {
		log.Panic(err)
	}
	line.Color = ColorYellow
	p.Add(line)
	vgCanvas := vgsvg.New(vg.Length(width), vg.Length(height))
	drawCanvas := draw.New(vgCanvas)
	drawCanvas = draw.Crop(drawCanvas, 0, 0, 0, -5) // prevents top y axis label from getting chopped
	p.Draw(drawCanvas)

	buf := new(bytes.Buffer)
	_, err = vgCanvas.WriteTo(buf)
	if err != nil {
		fmt.Println("could not write SVG", err)
	}
	return buf.String()
}
