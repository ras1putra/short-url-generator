package redirect

import (
	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
)

type AdGroups struct {
	Popup        []repository.Ad
	Banner       []repository.Ad
	Native       []repository.Ad
	Video        []repository.Ad
	Interstitial []repository.Ad
}

func GroupAds(ads []repository.Ad) AdGroups {
	var g AdGroups
	for _, ad := range ads {
		switch ad.AdType {
		case constants.AdTypePopup:
			g.Popup = append(g.Popup, ad)
		case constants.AdTypeBanner:
			g.Banner = append(g.Banner, ad)
		case constants.AdTypeNative:
			g.Native = append(g.Native, ad)
		case constants.AdTypeVideo:
			g.Video = append(g.Video, ad)
		case constants.AdTypeInterstitial:
			g.Interstitial = append(g.Interstitial, ad)
		}
	}
	return g
}
