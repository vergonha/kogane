package middleware

import "strings"

var socialPreviewBots = []string{
	"facebookexternalhit",
	"twitterbot",
	"discordbot",
	"whatsapp",
	"telegrambot",
	"slackbot",
	"linkedinbot",
	"skypeuripreview",
}

func IsSocialPreviewBot(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	for _, sig := range socialPreviewBots {
		if strings.Contains(ua, sig) {
			return true
		}
	}
	return false
}
