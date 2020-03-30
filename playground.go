package main

import "fmt"

type obj struct {
	names []string
}

type objList []obj

var ol = objList{
	{
		names: []string{"Jeremy"},
	},
}

func main() {
	var listLen = len(ol[0].names)
	fmt.Println(ol[0].names[listLen-1])
	ol[0].names = append(ol[0].names, "bob")
	listLen = len(ol[0].names)
	fmt.Println(ol[0].names[listLen-1])
}
