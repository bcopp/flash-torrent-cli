package main

import (
	"fmt"
	"testing"
)

func TestPirateBayURL(t *testing.T) {
	var url = createPirateURL("rick and morty")
	var goodUrl = "https://thepiratebay.org/search/rick%and%morty/0/99/0"
	if url != goodUrl {
		t.Errorf(fmt.Sprintf("Error\nURL was: %[1]v\nexpected: %[2]v", url, goodUrl))
	}
}
