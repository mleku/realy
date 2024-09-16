package normalize

import (
	"fmt"
	"testing"
)

func TestURL(t *testing.T) {
	fmt.Println(URL(B("")))
	fmt.Println(URL(B("wss://x.com/y")))
	fmt.Println(URL(B("wss://x.com/y/")))
	fmt.Println(URL(B("http://x.com/y")))
	fmt.Println(URL(URL(B("http://x.com/y"))))
	fmt.Println(URL(B("wss://x.com")))
	fmt.Println(URL(B("wss://x.com/")))
	fmt.Println(URL(URL(URL(B("wss://x.com/")))))
	fmt.Println(URL(B("x.com")))
	fmt.Println(URL(B("x.com/")))
	fmt.Println(URL(B("x.com////")))
	fmt.Println(URL(B("x.com/?x=23")))

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
