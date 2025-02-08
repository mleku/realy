package normalize

import (
	"fmt"
	"testing"
)

func TestURL(t *testing.T) {
	fmt.Println(URL([]byte("")))
	fmt.Println(URL([]byte("wss://x.com/y")))
	fmt.Println(URL([]byte("wss://x.com/y/")))
	fmt.Println(URL([]byte("http://x.com/y")))
	fmt.Println(URL(URL([]byte("http://x.com/y"))))
	fmt.Println(URL([]byte("wss://x.com")))
	fmt.Println(URL([]byte("wss://x.com/")))
	fmt.Println(URL(URL(URL([]byte("wss://x.com/")))))
	fmt.Println(URL([]byte("x.com")))
	fmt.Println(URL([]byte("x.com/")))
	fmt.Println(URL([]byte("x.com////")))
	fmt.Println(URL([]byte("x.com/?x=23")))

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
