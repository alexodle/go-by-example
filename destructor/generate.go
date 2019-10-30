package destructor

func GenerateWrappers(inputDir string, outputDir string) {
	structs := Parse(inputDir)
	files := Model(structs, inputDir, outputDir)
	WriteCode(files)
}
