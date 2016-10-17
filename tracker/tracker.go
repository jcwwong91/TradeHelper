package tracker

import(
//	"log"
	"fmt"

	"tradeHelper/stock"

//	"github.com/FlashBoys/go-finance"
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
func (t *Tracker) TrackStock(ticker string) error {

	if t.stocks[ticker] != nil {
		return fmt.Errorf("The stock with ticker %s is already tracked", ticker)
	}

	t.stocks[ticker] = &stock.Stock{
		Ticker: ticker,
	}
	//TODO: Add stock to DB
	//TODO: Fetch historical Data

	return nil
}

// GetStockConfig retrieves the technical analysis information/configuration
// about a particular stock
func (t *Tracker) GetStockConfig(ticker string) (*stock.Stock, error) {
	s := t.stocks[ticker]
	if s == nil {
		return nil, fmt.Errorf("The stock with ticker %s was not found", ticker)
	}
	return s, nil
}

// StoptrackingStock removes a particular stock from the list of tracked stocks
func (t *Tracker) StopTrackingStock(ticker string) error {
	if t.stocks[ticker] == nil {
		return fmt.Errorf("The stock with ticker %s was not found", ticker)
	}
	t.stocks[ticker] = nil
	return nil
}
