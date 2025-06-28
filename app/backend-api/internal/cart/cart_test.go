package cart

import (
	"math"
	"testing"
)

func TestBuildCart_Empty(t *testing.T) {
	c := buildCart(nil)
	if c.TotalItems != 0 || c.TotalPrice != 0 {
		t.Errorf("empty cart should have 0 items and 0 price; got %d items, %.2f price", c.TotalItems, c.TotalPrice)
	}
}

func TestBuildCart_SingleItem(t *testing.T) {
	items := []CartItem{
		{ProductID: "p1", Price: 25.00, Quantity: 2},
	}
	c := buildCart(items)
	if c.TotalItems != 2 {
		t.Errorf("TotalItems = %d; want 2", c.TotalItems)
	}
	if c.TotalPrice != 50.00 {
		t.Errorf("TotalPrice = %.2f; want 50.00", c.TotalPrice)
	}
}

func TestBuildCart_MultipleItems_RoundingToTwoDecimals(t *testing.T) {
	// 3 * 1.10 = 3.30, but floating-point could give 3.3000000000000003 — must round.
	items := []CartItem{
		{ProductID: "p1", Price: 1.10, Quantity: 3},
	}
	c := buildCart(items)
	want := math.Round(3*1.10*100) / 100
	if c.TotalPrice != want {
		t.Errorf("TotalPrice = %v; want %v", c.TotalPrice, want)
	}
}

func TestBuildCart_TotalItems_IsSumOfQuantities(t *testing.T) {
	items := []CartItem{
		{ProductID: "p1", Price: 10.00, Quantity: 2},
		{ProductID: "p2", Price: 5.00, Quantity: 3},
		{ProductID: "p3", Price: 20.00, Quantity: 1},
	}
	c := buildCart(items)
	if c.TotalItems != 6 {
		t.Errorf("TotalItems = %d; want 6", c.TotalItems)
	}
	wantPrice := math.Round((2*10.00+3*5.00+20.00)*100) / 100
	if c.TotalPrice != wantPrice {
		t.Errorf("TotalPrice = %.2f; want %.2f", c.TotalPrice, wantPrice)
	}
}
