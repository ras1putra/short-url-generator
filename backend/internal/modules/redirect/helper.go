package redirect

import "urlshortener/internal/repository"

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
		case "POPUP":
			g.Popup = append(g.Popup, ad)
		case "BANNER":
			g.Banner = append(g.Banner, ad)
		case "NATIVE":
			g.Native = append(g.Native, ad)
		case "VIDEO":
			g.Video = append(g.Video, ad)
		case "INTERSTITIAL":
			g.Interstitial = append(g.Interstitial, ad)
		}
	}
	return g
}
