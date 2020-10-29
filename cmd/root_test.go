package cmd

import "testing"

func TestAllFilter(t *testing.T){
	s := filterClusters("all")
	if len(s) != 7{
		t.Fatal("len not 7")
	}
}


func TestTestFilter(t *testing.T){
	s := filterClusters("test")
	if len(s) != 2{
		t.Fatal("len not 2")
	}
}

func TestStageFilter(t *testing.T){
	s := filterClusters("stage")
	if len(s) != 2{
		t.Fatal("len not 2")
	}
}

func TestProductFilter(t *testing.T){
	s := filterClusters("product")
	if len(s) != 3{
		t.Fatal("len not 3")
	}
}
