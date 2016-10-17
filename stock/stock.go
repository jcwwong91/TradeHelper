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


	bars, err := finance.GetQuoteHistory(s.Ticker, start, end, finance.IntervalDaily)
	if err != nil {
		log.Println("Error getting quote history for '%s'", s.Ticker)
	}

	var min, max, sum float64
	minimas := []float64{}
	maximas := []float64{}
	for i, bar := range bars {
		v,_ := bar.Close.Float64() // Don't care if exact
		if v < min || i == 0{
			min = v
		}
		if v > max {
			max = v
		}
		if i != 0 && i != len(bars) -1 {
			prev, _ := bars[i-1].Close.Float64()
			next, _ := bars[i+1].Close.Float64()
			if prev < v && next < v {
				maximas = append(maximas, v)
			}
			if prev> v && next > v {
				minimas = append(minimas, v)
			}
		}
		sum += v
	}

	// Calculate the standard deviation
	mean := sum/float64(len(bars))
	sum = 0
	for _, bar := range bars {
		v, _ := bar.Close.Float64()
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
