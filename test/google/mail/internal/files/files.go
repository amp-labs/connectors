package files

import _ "embed"

var (
	//go:embed bridge.png
	ImageBridge []byte
	//go:embed norway.png
	ImageNorway []byte
	//go:embed house.png
	ImageHouse []byte
	//go:embed "Develop on Google Workspace.pdf"
	DocumentAboutGoogleDev []byte
	//go:embed "html-attachment-test.html"
	DocumentHtml []byte
)
