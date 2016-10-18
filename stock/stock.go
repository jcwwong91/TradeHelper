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
	Peaks int		// Number of times hit resistance
	Troughs int		// Number of times hit support
	Min float64		// Assume this is support value
	Max float64		// Assume this is resistance value
	LastClose float64	// The last closing price detected
	LastOpen float64	// The last openning price detected

	sync.Mutex
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
	var min, max, sum float64
	minimas := []float64{}
	maximas := []float64{}
	dayAvgs := []float64{}
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
		if lower < min || i == 0{
			min = lower
		}
		if upper > max {
			max = upper
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
				maximas = append(maximas, upper)
			}
			if prevLow > lower && nextLow > lower {
				minimas = append(minimas, lower)
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

	s.info.Peaks = 0
	tolerance := sd * s.config.RSTolerance
	for _, v := range maximas {
		if v + tolerance >= max {
			s.info.Peaks++
		}
	}

	s.info.Troughs = 0
	for _, v:= range minimas {
		if v - tolerance <= min {
			s.info.Troughs++
		}
	}
	s.info.Min = min
	s.info.Max = max
}
