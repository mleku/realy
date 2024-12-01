package normalize

import (
	"fmt"
	"testing"
)

func TestURL(t *testing.T) {
	fmt.Println(URL(by("")))
	fmt.Println(URL(by("wss://x.com/y")))
	fmt.Println(URL(by("wss://x.com/y/")))
	fmt.Println(URL(by("http://x.com/y")))
	fmt.Println(URL(URL(by("http://x.com/y"))))
	fmt.Println(URL(by("wss://x.com")))
	fmt.Println(URL(by("wss://x.com/")))
	fmt.Println(URL(URL(URL(by("wss://x.com/")))))
	fmt.Println(URL(by("x.com")))
	fmt.Println(URL(by("x.com/")))
	fmt.Println(URL(by("x.com////")))
	fmt.Println(URL(by("x.com/?x=23")))

	// Output:
	//
	// wss://x.com/y
	// wss://x.com/y
	// ws://x.com/y
	// ws://x.com/y
	// wss://x.com
	// wss://x.com
	// wss://x.com
	// wss://x.com
	// wss://x.com
	// wss://x.com
	// wss://x.com?x=23
}
