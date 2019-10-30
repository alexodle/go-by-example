package main

import "fmt"

const (
	metadataServiceIP      = "169.254.169.254"
	metadataServiceVersion = "2019-06-04"
)

func main() {
	fmt.Println("hello world")

	url1 := fmt.Sprintf(`http://%s/metadata/instance/compute/resourceId?api-version=%s\&format=text`, metadataServiceIP, metadataServiceVersion)
	url2 := fmt.Sprintf("http://%s/metadata/instance/compute/resourceId?api-version=%s\\&format=text", metadataServiceIP, metadataServiceVersion)
	fmt.Println(url1 == url2)
}
