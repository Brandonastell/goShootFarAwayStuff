package preprocess

func read_config(filename string) []byte {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println(raw)
	return raw
}
