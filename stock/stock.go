package stock

import(
	"log"
	"math"
	"time"
	"sync"

	"github.com/FlashBoys/go-finance"
)

type Stock struct {
	Ticker string
	info Info
	config Config
}

type Config struct {
	RSTolerance float64
}

type Info struct {
	Min Point		// Minimum stock price found
	Max Point		// Maximum stock price found
	LastClose float64	// The last closing price detected
	LastOpen float64	// The last openning price detected
	Supports []*Trend	// List of trends found at support level
	Resistances []*Trend	// List of trends found at resistance level

	SMA []float64		// Simple moving averages for current stock
	EMA []float64		// Exponential moving averages for current stock

	sync.Mutex
}

type Point struct {
	T time.Time	// The time this point represents
	V float64	// The time this point represenets
}

// GetInfo returns a copy of some calculated information about the stock
func (s *Stock) GetInfo() Info {
	s.info.Lock()
	defer s.info.Unlock()
	return s.info
}

// GetConfig returns a copy of the stock's configuration
func (s *Stock) GetConfig() Config {
	return s.config
}

// SetConfig sets the configuration parameters this screener looks for
func (s *Stock) SetConfig(cfg Config) {
	s.config = cfg
}

// CalcualateInfo calculates various variables regarding a stock
func (s *Stock) CalculateInfo() {
	s.info.Lock()
	defer s.info.Unlock()
	now := time.Now()
	end := now
	start := end.AddDate(0, -3, 0)	// POC will limit to 3 months before


	// Bars return from most recent to oldest
	bars, err := finance.GetQuoteHistory(s.Ticker, start, end, finance.IntervalDaily)
	if err != nil {
		log.Println("Error getting quote history for '%s'", s.Ticker)
	}

	s.calcRS(bars)
	s.info.LastClose, _ = bars[0].Close.Float64()
	s.info.LastOpen, _ = bars[0].Open.Float64()
}

func (s *Stock) calcRS(bars []*finance.Bar) {
	var min, max Point
	minimas := []*Point{}
	maximas := []*Point{}
	dayAvgs := []float64{}
	var sum float64
	for i, bar := range bars {
		var upper, lower float64
		open ,_ := bar.Open.Float64() // Don't care if exact
		close ,_ := bar.Close.Float64() // Don't care if exact
		if open > close {
			upper = open
			lower = close
		} else {
			upper = close
			lower = open
		}
		if lower < min.V || i == 0{
			min = Point{V:lower, T:bar.Date}
		}
		if upper > max.V {
			max = Point{V:upper, T:bar.Date}
		}
		if i != 0 && i != len(bars) -1 {
			var prevLow, prevUp float64
			prevOpen, _ := bars[i-1].Open.Float64()
			prevClose, _ := bars[i-1].Close.Float64()
			if prevOpen > prevClose {
				prevUp = prevOpen
				prevLow = prevClose
			} else {
				prevUp = prevClose
				prevLow = prevOpen
			}

			var nextLow, nextUp float64
			nextOpen, _ := bars[i+1].Open.Float64()
			nextClose, _ := bars[i+1].Close.Float64()
			if nextOpen > nextClose {
				nextUp = nextOpen
				nextLow = nextClose
			} else {
				nextUp = nextClose
				nextLow = nextOpen
			}

			if prevUp < upper && nextUp < upper {
				maximas = append(maximas, &Point{V:upper, T:bar.Date})
			}
			if prevLow > lower && nextLow > lower {
				minimas = append(minimas, &Point{V:lower, T:bar.Date})
			}
		}
		dayAvg := (upper + lower)/2
		dayAvgs = append(dayAvgs, dayAvg)
		sum += dayAvg
	}

	// Calculate the standard deviation
	mean := sum/float64(len(bars))
	sum = 0
	for _, v := range dayAvgs {
		diff := v - mean
		sum += diff * diff
	}
	sd := math.Sqrt(sum)

	s.info.Supports = s.findTrends(minimas, sd*s.config.RSTolerance)
	s.info.Resistances = s.findTrends(maximas, sd*s.config.RSTolerance)
	s.calcSMA(bars)
	s.calcEMA(bars)
	s.info.Min = min
	s.info.Max = max
}

func (s *Stock) findTrends(points []*Point, tol float64) []*Trend{

	trends := []*Trend {}
	for i, _ := range points {
		v1 := points[i].V
		for j := i+1; j < len(points); j++ {
			v2 := points[j].V
			duration := points[i].T.Sub(points[j].T)
			trend := Trend{}
			trend.Slope = (v2 - v1) / duration.Seconds()
			trend.Constant = v2 - (trend.Slope * float64(points[j].T.Unix()))
			trend.Points = []*Point{}
			trends = append(trends, &trend)
		}
	}

	for _, t := range trends {
		for _, p := range points {
			v := (t.Slope * float64(p.T.Unix())) + t.Constant
			lower := v - tol
			upper := v + tol
			if p.V > lower && p.V < upper {
				t.Hits++
				t.Points = append(t.Points,p)
			}
		}
	}
	return trends
}

func (s *Stock) calcEMA(bars []*finance.Bar) {
	s.info.EMA = []float64{0}
	for i, bar := range bars {
		multiplier := (2.0 /  (float64(i) + 1.0))
		v,_ := bar.Close.Float64()
		ema := (v - s.info.EMA[i]) * multiplier + s.info.EMA[i]
		s.info.EMA = append(s.info.EMA, ema)
	}
	s.info.EMA[0] = 0.0 // Seems ot have +inf which messes with json
}

func (s *Stock) calcSMA(bars []*finance.Bar) {
	sum := 0.0
	s.info.SMA = []float64{}
	for i, bar := range bars {
		c, _ := bar.Close.Float64()
		sum += c
		s.info.SMA = append(s.info.SMA, sum/float64(i))
	}
	s.info.SMA[0] = 0.0	// Seems to have +inf which is messing with JSON
}
