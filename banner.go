package main

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgsvg"
	"image/color"
	"log"
	"strings"
	"text/template"
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

// BannerXY represents a single point on the plot.
type BannerXY struct {
	X time.Time
	Y int
}

type Theme struct {
	Background string `json:"background"`
	TextTicks string `json:"text_ticks"`
	CurrentBPM string `json:"current_bpm"`
	Title string `json:"title"`
	Heart string `json:"heart"`
	Axes string `json:"axes"`
}

type Template struct {
	Width int
	Height int
	PaddingTopBottom int
	Theme Theme

	Plot string

	Heart string
	BPM int
	BPMTextSize int

	Title string
	TitleSize int
	ShowWatermark bool
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

func defaultBanner(width, height int) string {
	bg := fmt.Sprintf(`<rect width="100%%" height="100%%" fill="%s" />`, HTMLCode(ColorCoffee))
	t := fmt.Sprintf(`<text x="%d" y="%d" fill="%s" style="font-family: sans-serif; font-weight:500;" dominant-baseline="hanging" text-anchor="middle">Banner not setup yet, or no data within range is available.</text>`, width/2, height/2, HTMLCode(ColorCream))
	banner := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" id="banner" width="%dpt" height="%dpt"> %s </svg>`, width, height, bg+t)
	return banner
}

func updateSVG(config Config) string {
	hrts, err := heartRateTimesSeries(&config)
	if err != nil {
		log.Print("Error grabbing time series", err.Error())
		return defaultBanner(500, 100)
	}
	banner, err := genBanner(hrts, config.DisplaySourceLink)
	if err != nil {
		log.Print("Error generating banner: ", err.Error())
		return defaultBanner(500, 100)
	}
	return banner
}

func genBanner(xy []BannerXY, showWatermark bool) (string, error) {
	timeSeries := make(plotter.XYs, 0, len(xy))
	for i := range xy {
		timeSeries = append(timeSeries, plotter.XY{
			X: float64(xy[i].X.Unix()),
			Y: float64(xy[i].Y),
		})
	}

	bpm := 0
	if len(timeSeries) <= 0 {
		return defaultBanner(500, 100), fmt.Errorf("data set empty")
	}
	bpm = int(timeSeries[len(timeSeries)-1].Y)

	bannerWidth := 500 // in pts
	bannerHeight := 100
	thirdWidth := bannerWidth / 3 // heart takes up 1/3rd, plot 2/3rd
	plotWidth := thirdWidth * 2

	tData := Template{
		Width:            500,
		Height:           100,
		PaddingTopBottom: 20,
		Theme: Theme{
			Background:     HTMLCode(ColorCoffee),
			TextTicks:       HTMLCode(ColorCream),
			CurrentBPM: HTMLCode(ColorCream),
			Title:      HTMLCode(ColorCream),
			Heart:          HTMLCode(ColorYellow),
			Axes:      HTMLCode(ColorOrange),
		},
		Plot:          genPlot(timeSeries, plotWidth, bannerHeight),
		Heart:         genHeart(bpm, thirdWidth),
		BPM:           bpm,
		BPMTextSize:   19,
		Title:         "Heart Rate from My FitBit Watch (Past 4 Hours)",
		TitleSize:     12,
		ShowWatermark: showWatermark,
	}

	t, err := template.New("banner").Funcs(sprig.GenericFuncMap()).Parse(tmplSVG)
	if err != nil {
		return "", err
	}
	b := new(bytes.Buffer)
	err = t.Execute(b, tData)
	if err != nil {
		return "", err
	}

	return b.String(), nil
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
	plotSVG := buf.String()

	plotSVG = fmt.Sprintf(`<g transform="translate(%d,%d)"> %s </g>`, 0, 0, plotSVG)
	plotSVG = strings.ReplaceAll(plotSVG, `font-family:Times;font-weight:normal;font-style:normal;font-size:10px;`, "") // remove in-line style
	plotSVG = strings.ReplaceAll(plotSVG, `<?xml version="1.0"?>`, "")                                                  // cannot have multiple xml tags
	plotSVG = strings.ReplaceAll(plotSVG, "<text", `<text class="text"`)
	return plotSVG
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

	heart = fmt.Sprintf(`<g transform="translate(%d %d)"> %s </g>`,0,-22,heart)
	return heart
}

// language=SVG
var tmplSVG = `
<svg xmlns="http://www.w3.org/2000/svg" id="banner" width="{{ .Width }}pt" height="{{add .Height .TitleSize .PaddingTopBottom }}pt">
	<!-- Generated via https://github.com/f0nkey/fitbit-readme-stats -->
	<rect width="100%" height="100%" fill="{{ .Theme.Background }}"/>
	<style> .text {font: 600 9px "Arial", Sans-Serif; fill: {{ .Theme.Background }};} </style>
	<g id="padding" transform="translate(0 {{ div .PaddingTopBottom 2 }})">
		<text id="title" dominant-baseline="hanging" text-anchor="middle" style="font: 600 12pt 'Arial', Sans-Serif; fill: {{ .Theme.Title }}" x="{{div .Width 2}}pt"> 
			{{.Title}}
		</text>
		{{ if .ShowWatermark }}
			<a href="https://github.com/f0nkey/fitbit-readme-stats">
				<text id="title" dominant-baseline="hanging" style="font: 600 8pt 'Arial', Sans-Serif; fill: {{ .Theme.Title }};" x="5pt">View on GitHub</text>
			</a>
		{{ end }}
		<g id="main-content" transform="translate(0 {{ add .TitleSize 6 }})">
			<g id="plot" transform="translate(166,0)">
				<!-- Generated by SVGo and Plotinum VG -->
				{{.Plot}}
			</g>
			<g id="heart">
				{{ .Heart }}
			</g>
			<g id="heart-text" transform="translate( {{ $WidthBy3 := div .Width 3 }} {{ div $WidthBy3 2 }} {{ div .Height 2 }})">
				<text id="current-bpm-text" class="text" text-anchor="middle" x="0" y="79">Current BPM</text>
				<text id="bpm-number" class="text" dominant-baseline="middle" text-anchor="middle" x="0" y="0">{{ .BPM }}</text>
				<style> #current-bpm-text {font-size: {{ .BPMTextSize }}pt; fill: {{ .Theme.CurrentBPM}};}  #bpm-number {font-size: 35px; fill: {{ .Theme.Background }};}</style>
			</g>
		</g>
	</g>
</svg>`
