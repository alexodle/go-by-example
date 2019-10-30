package main

import "fmt"

func main() {
	s := make([]string, 3, 4)
	s[0] = "a"
	s[1] = "b"
	s[2] = "c"
	s2 := append(s, "d")
	s3 := append(s2, "e")

	fmt.Println("s:", s)
	fmt.Println("s2:", s2)
	fmt.Println("s3:", s3)
	fmt.Println("")

	s2[0] = "f"
	fmt.Println("s:", s)
	fmt.Println("s2:", s2)
	fmt.Println("s3:", s3)
	fmt.Println("")

	s3[0] = "g"
	fmt.Println("s:", s)
	fmt.Println("s2:", s2)
	fmt.Println("s3:", s3)
	fmt.Println("")
}
