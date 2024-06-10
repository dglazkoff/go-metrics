package main

func main() {
	parseFlags()

	if err := Run(); err != nil {
		panic(err)
	}
}
