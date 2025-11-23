package main

import "testing"

func TestSeven(t *testing.T) {
    var dir ="A:\\GoLang\\GoLang_Project\\demo\\resources"

	err := invoke7zip(dir, false )
	if err != nil {
		t.Error(err)
	}

}