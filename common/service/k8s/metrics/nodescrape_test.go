package metrics

import "testing"

func TestNodeScrape_Scrape(t *testing.T) {
	n := NewNodeScrape("218.23.2.55", "31992")
	data, err := n.Scrape()
	if err != nil {
		t.Errorf("Scrape() error = %v, wantErr %v", err, nil)
	}
	t.Logf("data: %v", data)
}
