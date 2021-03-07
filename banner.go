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
	"strconv"
	"strings"
	"text/template"
	"time"
)

// BannerXY represents a single point on the plot.
type BannerXY struct {
	X time.Time
	Y int
}

type Theme struct {
	Background  string `json:"background"`
	TextTicks   string `json:"text_ticks"`
	CurrentBPM  string `json:"current_bpm"`
	Title       string `json:"title"`
	Heart       string `json:"heart"`
	Axes        string `json:"axes"`
	PlotLine    string `json:"plot_line"`
	HeartNumber string `json:"heart_number"`
}

type Template struct {
	Width            int
	Height           int
	PaddingTopBottom int
	Theme            Theme

	Plot string

	Heart       string
	BPM         int
	BPMTextSize int

	Title         string
	TitleSize     int
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

func defaultBanner(c Config) string {
	bg := fmt.Sprintf(`<rect width="100%%" height="100%%" fill="%s" />`, c.Theme.Background)
	t := fmt.Sprintf(`<text x="%d" y="%d" fill="%s" style="font-family: sans-serif; font-weight:500;" dominant-baseline="hanging" text-anchor="middle">Banner not setup yet, or no data within range is available.</text>`, c.BannerWidth/2, c.BannerHeight/2, c.Theme.Title)
	banner := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" id="banner" width="%dpt" height="%dpt"> %s </svg>`, c.BannerWidth, c.BannerHeight, bg+t)
	return banner
}

func updateSVG(c *Config) string {
	hrts, err := heartRateTimesSeries(c)
	if err != nil {
		log.Print("Error grabbing time series", err.Error())
		return defaultBanner(*c)
	}
	banner, err := genBanner(hrts, *c)
	if err != nil {
		log.Print("Error generating banner: ", err.Error())
		return defaultBanner(*c)
	}
	return banner
}

func genBanner(xy []BannerXY, config Config) (string, error) {
	timeSeries := make(plotter.XYs, 0, len(xy))
	for i := range xy {
		timeSeries = append(timeSeries, plotter.XY{
			X: float64(xy[i].X.Unix()),
			Y: float64(xy[i].Y),
		})
	}

	bpm := 0
	if len(timeSeries) <= 0 {
		return defaultBanner(config), fmt.Errorf("data set empty")
	}
	bpm = int(timeSeries[len(timeSeries)-1].Y)

	thirdWidth := config.BannerWidth / 3 // heart takes up 1/3rd, plot 2/3rd
	plotWidth := thirdWidth * 2

	tData := Template{
		Width:            config.BannerWidth,
		Height:           config.BannerHeight,
		PaddingTopBottom: 20,
		Theme:            config.Theme,
		Plot:             genPlot(timeSeries, plotWidth, config),
		Heart:            genHeart(bpm, thirdWidth, config.Theme.Heart),
		BPM:              bpm,
		BPMTextSize:      19,
		Title:            config.BannerTitle,
		TitleSize:        12,
		ShowWatermark:    config.DisplayViewOnGitHub,
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

// RGBAFromString parses a color.RGBA from a string e.g. rgba(255,20,147,100).
func RGBAFromString(s string) color.RGBA {
	nums := s[strings.Index(s, "(")+1 : strings.Index(s, ")")]
	split := strings.Split(nums, ",")
	if len(split) != 4 {
		log.Println("Invalid theme color:", s)
		return color.RGBA{}
	}
	r, err := strconv.Atoi(strings.TrimSpace(split[0]))
	if err != nil {
		log.Println("Error converting theme color r component:", s)
		return color.RGBA{}
	}
	g, err := strconv.Atoi(strings.TrimSpace(split[1]))
	if err != nil {
		log.Println("Error converting theme color g component:", s)
		return color.RGBA{}
	}
	b, err := strconv.Atoi(strings.TrimSpace(split[2]))
	if err != nil {
		log.Println("Error converting theme color b component:", s)
		return color.RGBA{}
	}
	a, err := strconv.Atoi(strings.TrimSpace(split[3]))
	if err != nil {
		log.Println("Error converting theme color a component:", s)
		return color.RGBA{}
	}
	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(a),
	}
}

func genPlot(timeSeries plotter.XYs, width int, config Config) string {
	p,_ := plot.New()

	p.X.Tick.Marker = plot.TimeTicks{
		Ticker: BannerTicker(timeSeries),
		Format: "15:04",
		Time:   nil,
	}

	p.X.Tick.LineStyle.Color = RGBAFromString(config.Theme.Axes)
	p.X.LineStyle.Color = RGBAFromString(config.Theme.Axes)
	p.Y.Tick.LineStyle.Color = RGBAFromString(config.Theme.Axes)
	p.Y.LineStyle.Color = RGBAFromString(config.Theme.Axes)

	p.Y.Label.TextStyle.Color = RGBAFromString(config.Theme.TextTicks)
	p.X.Label.TextStyle.Color = RGBAFromString(config.Theme.TextTicks)
	p.X.Tick.Label.Color = RGBAFromString(config.Theme.TextTicks)
	p.Y.Tick.Label.Color = RGBAFromString(config.Theme.TextTicks)

	p.BackgroundColor = RGBAFromString(config.Theme.Background)

	line, err := plotter.NewLine(timeSeries)
	if err != nil {
		log.Panic(err)
	}
	line.Color = RGBAFromString(config.Theme.PlotLine)
	p.Add(line)
	vgCanvas := vgsvg.New(vg.Length(width), vg.Length(config.BannerHeight))
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

func genHeart(bpm int, width int, heartColor string) string {
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
	`, width, width, viewBox, viewBox, gOffset, gOffset, heartColor, 60000/bpm)

	heart = fmt.Sprintf(`<g transform="translate(%d %d)"> %s </g>`, 0, -22, heart)
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
				<style> #current-bpm-text {font-size: {{ .BPMTextSize }}pt; fill: {{ .Theme.CurrentBPM}};}  #bpm-number {font-size: 35px; fill: {{ .Theme.HeartNumber }};}</style>
			</g>
		</g>
	</g>
</svg>`
