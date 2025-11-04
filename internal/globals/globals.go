package globals

var codecraftersServerURL string

func SetCodecraftersServerURL(url string) {
	codecraftersServerURL = url
}

func GetCodecraftersServerURL() string {
	if codecraftersServerURL == "" {
		panic("CodecraftersServerURL is not set")
	}
	return codecraftersServerURL
}

