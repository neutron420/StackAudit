package output

import (
	"fmt"
	"os"
)

const bannerText = `
   _____ _             _      _             _ _ _ 
  / ____| |           | |    / \           | (_) |
 | (___ | |_ __ _  ___| | __/ _ \ _   _  __| |_| |_ 
  \___ \| __/ _' |/ __| |/ / /_\ \ | | |/ _' | | __|
  ____) | || (_| | (__|   < / ___ \ |_| | (_| | | |_ 
 |_____/ \__\__,_|\___|_|\_\_/   \_\__,_|\__,_|_|\__|
                                                     
      Production Health & Security Audit Tool
`

func PrintBanner() {
	if os.Getenv("TERM") == "dumb" {
		fmt.Println("StackAudit - Production Health & Security Audit Tool")
		return
	}

	banner := styleBranding.Render(bannerText)
	fmt.Println(banner)
}
