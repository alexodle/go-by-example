package destructor

func GenerateWrappers(inputDir string, outputDir string) {
	structs := Parse(inputDir)
	files := Remodel(structs, inputDir, outputDir)
	WriteCode(files)
}
