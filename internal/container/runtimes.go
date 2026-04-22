package container

// RuntimeInfo contains information about a supported function runtime env.
type RuntimeInfo struct {
	Image                Image
	InvocationCmd        []string
	ConcurrencySupported bool
}

type Image struct {
	LocalImage  string
	RemoteImage string
}

const CUSTOM_RUNTIME = "custom"

var refreshedImages = map[string]bool{}

var RuntimeToInfo = map[string]RuntimeInfo{
	"python310": {
		Image{
			"localhost:5000/python310",
			"grussorusso/serverledge-python310",
		},
		[]string{"python", "/entrypoint.py"},
		true},
	"nodejs17ng": {
		Image{
			"localhost:5000/nodejs17ng",
			"grussorusso/serverledge-nodejs17ng",
		},
		[]string{},
		false},
}
