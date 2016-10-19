package tracker

import(
	"bufio"
	"fmt"
	"os"

	"tradeHelper/stock"
)

type Tracker struct{
	stocks map[string]*stock.Stock
}

func NewTracker() *Tracker{
	// TODO: Load old state from DB
	tracker := &Tracker{
		stocks: make(map[string]*stock.Stock),
	}
	return tracker
}

// TrackStock adds a particular stock to a list of stocks to be tracked by the
// application
func (t *Tracker) TrackStock(ticker string, tolerance float64) error {

	if t.stocks[ticker] != nil {
		return fmt.Errorf("The stock with ticker %s is already tracked", ticker)
	}

	s := &stock.Stock{
		Ticker: ticker,
	}

	cfg := stock.Config {
		RSTolerance: tolerance,
	}
	s.SetConfig(cfg)

	//TODO: Add stock to DB	
	t.stocks[ticker] = s
	go func() {
		s.CalculateInfo()
	}()

	return nil
}

// GetTrackedStocks returns a list of all the stocks it is currently tracking
func (t *Tracker) GetTrackedStocks() ([]string) {
	ret := []string{}
	for k, _ := range t.stocks {
		ret = append(ret, k)
	}
	return ret
}

// GetStockConfig retrieves the technical analysis information/configuration
// about a particular stock
func (t *Tracker) GetStockConfig(ticker string) (*stock.Config, error) {
	s := t.stocks[ticker]
	if s == nil {
		return nil, fmt.Errorf("The stock with ticker %s was not found", ticker)
	}
	config := s.GetConfig()
	return &config, nil
}

// StoptrackingStock removes a particular stock from the list of tracked stocks
func (t *Tracker) StopTrackingStock(ticker string) error {
	if t.stocks[ticker] == nil {
		return fmt.Errorf("The stock with ticker %s was not found", ticker)
	}
	t.stocks[ticker] = nil
	return nil
}

// GetStockInfo returns a list of information regarding a particular stock
func (t *Tracker) GetStockInfo(ticker string) (*stock.Info, error) {
	if t.stocks[ticker] == nil {
		return nil, fmt.Errorf("The stock with ticker %s was not found", ticker)
	}
	info := t.stocks[ticker].GetInfo()
	return &info, nil
}

// Load takes in a csv and fetchs all stocks specified in the file
func (t *Tracker) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t.TrackStock(scanner.Text(),0.15)
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
