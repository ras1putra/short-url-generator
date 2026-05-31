package redirect

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"urlshortener/internal/repository"
	"urlshortener/pkg/constants"
)

func RenderInterstitial(ads []repository.Ad, url repository.Url, bridgeToken string, primaryAdID uuid.UUID) string {
	targetURL := url.Original
	g := GroupAds(ads)

	var popupHTML string
	if len(g.Popup) > 0 {
		popupHTML = popupOverlay(&g.Popup[0], url.Slug)
	} else {
		popupHTML = popupPlaceholder()
	}

	var sections []string
	if len(g.Banner) > 0 {
		sections = append(sections, bannerStrip(&g.Banner[0], url.Slug))
	} else {
		sections = append(sections, bannerPlaceholder())
	}

	if len(g.Native) > 0 {
		sections = append(sections, fmt.Sprintf(`<div class="native-grid">%s</div>`, nativeCard(&g.Native[0], url.Slug)))
	} else {
		sections = append(sections, fmt.Sprintf(`<div class="native-grid">%s</div>`, nativePlaceholder()))
	}

	if len(g.Video) > 0 {
		sections = append(sections, videoSection(&g.Video[0], url.Slug))
	} else {
		sections = append(sections, videoPlaceholder())
	}

	if len(g.Interstitial) > 0 {
		sections = append(sections, interstSection(&g.Interstitial[0], url.Slug))
	} else {
		sections = append(sections, interstPlaceholder())
	}

	body := strings.Join(sections, "\n")
	token := bridgeToken

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Redirecting...</title>
<link href="https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700&display=swap" rel="stylesheet">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:'DM Sans',sans-serif;background:#f8fafc;color:#0f172a;height:100vh;overflow:hidden;display:flex;flex-direction:column;align-items:center;padding:0}

/* ── Popup overlay ── */
#popup-overlay{position:fixed;inset:0;z-index:1000;background:rgba(15,23,42,.35);backdrop-filter:blur(6px);display:flex;align-items:center;justify-content:center;padding:16px}
#popup-overlay.hidden{display:none}
.popup-card{position:relative;height:90vh;width:min(100vw,calc(90vh*5/6));max-width:100vw;aspect-ratio:5/6;background:#ffffff;border-radius:20px;overflow:hidden;box-shadow:0 20px 50px rgba(0,0,0,.15),0 0 0 1px rgba(0,0,0,.05);display:flex;align-items:center;justify-content:center}
.popup-card .popup-countdown{position:absolute;top:14px;left:14px;z-index:10;color:#0f172a;font-size:12px;font-weight:700;letter-spacing:.04em;background:rgba(255,255,255,.95);padding:6px 10px;border-radius:999px;backdrop-filter:blur(8px);box-shadow:0 2px 10px rgba(0,0,0,.08);border:1px solid rgba(0,0,0,.06);font-variant-numeric:tabular-nums}
.popup-card .popup-progress{position:absolute;left:14px;right:14px;top:48px;height:4px;background:rgba(255,255,255,.5);border-radius:999px;overflow:hidden;z-index:10}
.popup-card .popup-progress .popup-progress-bar{height:100%%;width:100%%;background:linear-gradient(90deg,#e5e7eb,#9ca3af);transform-origin:left center;transform:scaleX(1);transition:transform 1s linear}
.popup-card .popup-close{position:absolute;top:12px;right:12px;z-index:10;width:32px;height:32px;border-radius:50%%;border:none;background:rgba(0,0,0,.05);color:#475569;font-size:18px;cursor:pointer;display:flex;align-items:center;justify-content:center;line-height:1;transition:background .2s,color .2s;backdrop-filter:blur(8px)}
.popup-card .popup-close:hover{background:rgba(0,0,0,.12);color:#0f172a}
.popup-card .popup-media{display:block;width:100%%;height:100%%}
.popup-card .popup-media img,.popup-card .popup-media video{width:100%%;height:100%%;object-fit:cover}

/* ── Main layout ── */
#main-content{width:100%%;max-width:1100px;height:100vh;padding:16px;display:grid;grid-template-columns:1fr;grid-template-rows:auto 1fr 1fr auto;gap:12px;margin:0 auto}
@media(min-width:960px){
  #main-content{grid-template-columns:1.1fr .9fr;grid-template-rows:auto 1fr auto;align-content:stretch}
}

/* ── Banner ── */
.banner-strip{width:100%%;background:#ffffff;border-radius:14px;overflow:hidden;border:1px solid rgba(0,0,0,.06);box-shadow:0 4px 20px -2px rgba(0,0,0,.05)}
.banner-strip a{display:block;width:100%%}
.banner-strip img,.banner-strip video{width:100%%;height:96px;object-fit:cover;display:block}
@media(min-width:960px){.banner-strip{grid-column:1 / -1}}

/* ── Native grid ── */
.native-grid{display:grid;grid-template-columns:1fr;gap:12px;min-height:0}
.native-card{background:#ffffff;border-radius:16px;border:1px solid rgba(0,0,0,.06);box-shadow:0 4px 20px -2px rgba(0,0,0,.05);overflow:hidden;transition:border-color .2s,transform .2s,box-shadow .2s}
.native-card:hover{border-color:rgba(75,85,99,.35);transform:translateY(-2px);box-shadow:0 10px 25px -5px rgba(0,0,0,.12),0 8px 16px -6px rgba(0,0,0,.08)}
.native-card .media{width:100%%;background:#f8fafc;display:flex;align-items:center;justify-content:center;height:180px;overflow:hidden;border-bottom:1px solid rgba(0,0,0,.04)}
.native-card .media img,.native-card .media video{width:100%%;height:100%%;object-fit:cover}
.native-card .body{padding:14px;display:flex;flex-direction:column;gap:8px}
.native-card .body .title{font-size:15px;font-weight:600;line-height:1.3;color:#0f172a}
.native-card .body .desc{font-size:13px;color:#475569;line-height:1.5}
.native-card .body .btn{display:inline-flex;align-items:center;gap:4px;padding:8px 16px;border-radius:10px;font-size:12px;font-weight:600;text-decoration:none;background:linear-gradient(135deg,#4b5563,#111827);color:#fff;align-self:flex-start;box-shadow:0 4px 12px rgba(0,0,0,.22);transition:opacity .2s}
.native-card .body .btn:hover{opacity:.9}

/* ── Video ── */
.video-section{width:100%%;background:#ffffff;border-radius:16px;border:1px solid rgba(0,0,0,.06);box-shadow:0 4px 20px -2px rgba(0,0,0,.05);overflow:hidden;min-height:0}
.video-section video{width:100%%;display:block;height:220px;object-fit:cover}
.video-section .body{padding:14px;display:flex;flex-direction:column;gap:8px}
.video-section .body .title{font-size:15px;font-weight:600;color:#0f172a}
.video-section .body .btn{display:inline-flex;align-items:center;gap:4px;padding:8px 16px;border-radius:10px;font-size:12px;font-weight:600;text-decoration:none;background:linear-gradient(135deg,#4b5563,#111827);color:#fff;align-self:flex-start;box-shadow:0 4px 12px rgba(0,0,0,.22);transition:opacity .2s}
.video-section .body .btn:hover{opacity:.9}

/* ── Interstitial ── */
.interst-section{width:100%%;background:#ffffff;border-radius:16px;border:1px solid rgba(0,0,0,.06);box-shadow:0 4px 20px -2px rgba(0,0,0,.05);overflow:hidden;position:relative;min-height:0;display:flex;align-items:center;justify-content:center}
.interst-section img,.interst-section video{position:absolute;inset:0;width:100%%;height:100%%;object-fit:cover}
.interst-section .overlay{position:relative;z-index:1;padding:40px 24px;text-align:center;background:linear-gradient(to top,rgba(15,23,42,.95),rgba(15,23,42,.45));width:100%%}
.interst-section .overlay .title{font-size:22px;font-weight:700;color:#fff;margin-bottom:6px}
.interst-section .overlay .desc{font-size:14px;color:rgba(255,255,255,.75);margin-bottom:16px}
.interst-section .overlay .btn{display:inline-flex;align-items:center;gap:4px;padding:10px 24px;border-radius:12px;font-size:14px;font-weight:600;text-decoration:none;background:#fff;color:#0f172a;transition:background .2s}
.interst-section .overlay .btn:hover{background:#f1f5f9}

/* ── Footer countdown (app-style) ── */
.footer{width:100%%;display:flex;align-items:center;justify-content:space-between;gap:16px;padding-top:4px}
@media(min-width:960px){.footer{grid-column:1 / -1}}
.countdown-widget{display:flex;align-items:center;gap:12px}
.ring-wrap{position:relative;width:52px;height:52px;flex-shrink:0}
.ring-wrap svg{transform:rotate(-90deg)}
.ring-bg{fill:none;stroke:rgba(0,0,0,.06);stroke-width:3.5}
.ring-fg{fill:none;stroke:url(#footerGrad);stroke-width:3.5;stroke-linecap:round;stroke-dasharray:138.2;stroke-dashoffset:0;transition:stroke-dashoffset 1s linear}
.ring-num{position:absolute;inset:0;display:flex;align-items:center;justify-content:center;font-size:16px;font-weight:700;font-variant-numeric:tabular-nums;color:#0f172a}
.countdown-text{display:flex;flex-direction:column;gap:2px}
.countdown-text .label{font-size:11px;font-weight:600;color:#64748b;letter-spacing:.06em;text-transform:uppercase}
.countdown-text .sublabel{font-size:13px;font-weight:500;color:#475569}
.skip-btn{background:rgba(0,0,0,.03);border:1px solid rgba(0,0,0,.06);color:#475569;font-size:13px;font-weight:500;cursor:pointer;padding:10px 20px;border-radius:12px;transition:all .2s;font-family:inherit;display:flex;align-items:center;gap:6px}
.skip-btn:hover:not(:disabled){background:rgba(75,85,99,.12);border-color:rgba(75,85,99,.3);color:#1f2937}
.skip-btn:disabled{opacity:.4;cursor:not-allowed;background:rgba(0,0,0,.02);color:#94a3b8;border-color:rgba(0,0,0,.04)}
.skip-btn .arrow{font-size:15px;transition:transform .2s}
.skip-btn:hover:not(:disabled) .arrow{transform:translateX(3px)}

/* ── Anti-adblock wall ── */
#adblock-wall{display:none;position:fixed;inset:0;z-index:9999;background:rgba(15,23,42,.95);backdrop-filter:blur(12px);align-items:center;justify-content:center;padding:24px;text-align:center;color:#fff;font-family:'DM Sans',sans-serif}
#adblock-wall.show{display:flex}
#adblock-wall .wall-content{max-width:420px}
#adblock-wall h2{font-size:22px;font-weight:700;margin-bottom:12px}
#adblock-wall p{font-size:14px;color:rgba(255,255,255,.7);margin-bottom:24px;line-height:1.6}
#adblock-wall .wall-icon{font-size:48px;margin-bottom:16px}
</style>
</head>
<body>
<div id="adblock-wall"><div class="wall-content"><div class="wall-icon">&#128274;</div><h2>Ad Blocker Detected</h2><p>Please disable your ad blocker to continue. This helps keep content free for everyone.</p></div></div>
%s
<div id="main-content">
%s
  <div class="footer">
    <div class="countdown-widget">
      <div class="ring-wrap">
        <svg width="52" height="52" viewBox="0 0 52 52">
          <defs>
            <linearGradient id="footerGrad" x1="0" y1="0" x2="1" y2="1">
            <stop offset="0%%" stop-color="#6b7280"/>
            <stop offset="100%%" stop-color="#111827"/>
            </linearGradient>
          </defs>
          <circle class="ring-bg" cx="26" cy="26" r="22"/>
          <circle class="ring-fg" id="footer-ring" cx="26" cy="26" r="22"/>
        </svg>
        <div class="ring-num" id="countdown">15</div>
      </div>
      <div class="countdown-text">
        <span class="label">Continue in</span>
        <span class="sublabel" id="countdown-sub">Please wait...</span>
      </div>
    </div>
    <button class="skip-btn" id="skip-btn" disabled onclick="skipAd()">
      Skip <span class="arrow">→</span>
    </button>
  </div>
</div>

<button id="honeypot-btn" style="position:absolute;left:-9999px;top:-9999px;width:1px;height:1px;opacity:.01" onclick="void(0)"></button>

<script>
(function(){
  var wall=document.getElementById('adblock-wall');
  var blocked=false;
  var mm=0,honeypotHit=false;

  document.addEventListener('mousemove',function(){mm++});
  document.getElementById('honeypot-btn').addEventListener('click',function(){honeypotHit=true});

  function getFP(){
    var parts=[screen.width+'x'+screen.height,screen.colorDepth,navigator.language,navigator.platform,new Date().getTimezoneOffset()];
    try{
      var c=document.createElement('canvas');c.width=200;c.height=50;
      var cx=c.getContext('2d');cx.textBaseline='top';cx.font='14px Arial';
      cx.fillText('fp'+navigator.userAgent.length,2,2);cx.strokeStyle='#f60';
      cx.strokeRect(0,0,50,50);parts.push(c.toDataURL().length);
    }catch(e){}
    var s=parts.join('|'),h=0;
    for(var i=0;i<s.length;i++){h=((h<<5)-h)+s.charCodeAt(i);h|=0}
    return h.toString(36);
  }
  var fp=getFP();

  var adScript = document.createElement('script');
  adScript.src = 'https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js';
  adScript.async = true;
  adScript.onerror = function(){blocked=true};
  document.head.appendChild(adScript);

  function postComplete(cb){
    var x=new XMLHttpRequest();
    x.open('POST','/api/r/'+encodeURIComponent(bridgeToken.split(':')[0])+'/complete?token='+encodeURIComponent(bridgeToken),true);
    x.setRequestHeader('Content-Type','application/json');
    x.onload=function(){
      if(x.status===200){
        var r=JSON.parse(x.responseText);
        if(r.success){window.location.href=r.destination_url;return}
      }
      if(cb)cb();
    };
    x.onerror=function(){if(cb)cb()};
    x.send(JSON.stringify({fingerprint:fp,honeypot_hit:honeypotHit,mouse_moves:mm}));
  }

  var popupTotal=10,ps=popupTotal,pc=document.getElementById('popup-countdown'),pb=document.getElementById('popup-close-btn'),ppb=document.getElementById('popup-progress-bar');
  if(pc){pc.textContent='Skip in '+ps+'s'}
  if(pc){var pi=setInterval(function(){ps--;if(pc)pc.textContent='Skip in '+Math.max(ps,0)+'s';if(ppb){var pct=Math.max(ps,0)/popupTotal;ppb.style.transform='scaleX('+pct+')'}if(ps<=0){clearInterval(pi);if(pc)pc.style.display='none';if(pb)pb.style.display='flex'}},1000)}
  function closePopup(){
    if(blocked)return;
    var overlay=document.getElementById('popup-overlay');
    if(overlay)overlay.classList.add('hidden');
    startMainCountdown();
  }
  window.closePopup=closePopup;

  var total=15,t=total;
  var ring=document.getElementById('footer-ring');
  var timer=document.getElementById('countdown');
  var sub=document.getElementById('countdown-sub');
  var skipBtn=document.getElementById('skip-btn');
  var targetURL=%q;
  var bridgeToken=%q;
  var circ=2*Math.PI*22;
  ring.style.strokeDasharray=circ;
  ring.style.strokeDashoffset=0;

  var subs=['Please wait...','Almost there...','Loading destination...','Hang tight!'];
  var subIdx=0;

  var cd=null;
  var countdownStarted=false;
  function startMainCountdown(){
    if(blocked)return;
    if(countdownStarted)return;
    countdownStarted=true;
    if(sub)sub.textContent=subs[0];
    cd=setInterval(function(){
      if(blocked){if(cd)clearInterval(cd);return}
      t--;
      var pct=t/total;
      ring.style.strokeDashoffset=circ*(1-pct);
      if(timer)timer.textContent=t;
      if(t<=5&&sub){subIdx=Math.min(subIdx+1,subs.length-1);sub.textContent=subs[subIdx]}
      if(t<=0){
        clearInterval(cd);
        if(timer)timer.textContent='…';
        if(sub)sub.textContent='Verifying...';
        if(skipBtn)skipBtn.disabled=true;
        postComplete(function(){
          if(timer)timer.textContent='!';
          if(sub)sub.textContent='Redirecting...';
          if(skipBtn)skipBtn.disabled=false;
          skipBtn.onclick=function(){window.location.href=targetURL};
        });
      }
    },1000);
  }

  if(!document.getElementById('popup-overlay')){
    startMainCountdown();
  }

  function skipAd(){
    if(blocked)return;
    if(!countdownStarted)return;
    if(t>0)return;
    if(cd)clearInterval(cd);
    if(timer)timer.textContent='…';
    if(sub)sub.textContent='Verifying...';
    postComplete(function(){window.location.href=targetURL});
  }
})();
</script>
</body>
</html>`, popupHTML, body, targetURL, token)
}

func adMediaTag(url, adType string) string {
	isVideo := strings.HasPrefix(adType, constants.AdTypeVideo) || strings.HasSuffix(url, ".mp4") || strings.HasSuffix(url, ".webm")
	if isVideo {
		return fmt.Sprintf(`<video src="%s" autoplay muted loop playsinline webkit-playsinline></video>`, url)
	}
	return fmt.Sprintf(`<img src="%s" alt="Ad" loading="eager">`, url)
}

func popupOverlay(ad *repository.Ad, slug string) string {
	adID := ad.ID.String()
	media := adMediaTag(ad.ImageUrl, ad.AdType)
return fmt.Sprintf(`<div id="popup-overlay">
  <div class="popup-card">
    <div class="popup-countdown" id="popup-countdown">Skip in 10s</div>
    <div class="popup-progress"><div class="popup-progress-bar" id="popup-progress-bar"></div></div>
    <button class="popup-close" id="popup-close-btn" onclick="closePopup()" style="display:none">&times;</button>
    <a href="/api/r/%s/click/%s" target="_blank" rel="noopener" class="popup-media">%s</a>
  </div>
</div>`, slug, adID, media)
}

func popupPlaceholder() string {
	return `<div id="popup-overlay">
  <div class="popup-card" style="flex-direction:column;align-items:stretch;justify-content:flex-end;padding:0;text-align:left;background:radial-gradient(140%% 100%% at 0%% 0%%,#1f2937 0%%,#111827 45%%,#030712 100%%);color:#fff">
    <div class="popup-countdown" id="popup-countdown">Skip in 10s</div>
    <div class="popup-progress"><div class="popup-progress-bar" id="popup-progress-bar"></div></div>
    <button class="popup-close" id="popup-close-btn" onclick="closePopup()" style="display:none">&times;</button>
    <div style="position:absolute;inset:0;background:linear-gradient(160deg,rgba(255,255,255,.06),rgba(17,24,39,.35) 35%%,rgba(3,7,18,.78));pointer-events:none"></div>
    <div style="position:relative;padding:22px;background:linear-gradient(to top,rgba(2,6,23,.96),rgba(2,6,23,.45));backdrop-filter:blur(2px);border-top:1px solid rgba(255,255,255,.08)">
      <div style="display:inline-flex;align-items:center;gap:6px;padding:4px 8px;border-radius:999px;background:rgba(255,255,255,.08);border:1px solid rgba(255,255,255,.15);font-size:11px;font-weight:700;letter-spacing:.06em;text-transform:uppercase;color:#d1d5db;margin-bottom:10px">Sponsored</div>
      <div style="font-size:24px;font-weight:800;line-height:1.2;color:#f8fafc;margin-bottom:8px;font-family:'DM Sans',sans-serif">Boost Your Reach Instantly</div>
      <div style="font-size:14px;line-height:1.6;color:rgba(226,232,240,.9);margin-bottom:16px;font-family:'DM Sans',sans-serif">Your campaign could be right here with full-screen impact and premium placement.</div>
      <a href="/" style="display:inline-flex;align-items:center;gap:8px;padding:11px 18px;border-radius:12px;font-size:13px;font-weight:700;text-decoration:none;background:linear-gradient(135deg,#f3f4f6,#d1d5db);color:#111827;box-shadow:0 6px 18px rgba(0,0,0,.22);font-family:'DM Sans',sans-serif" target="_blank" rel="noopener">Start Advertising <span style="font-size:16px;line-height:1">→</span></a>
    </div>
  </div>
</div>`
}

func bannerStrip(ad *repository.Ad, slug string) string {
	adID := ad.ID.String()
	media := adMediaTag(ad.ImageUrl, ad.AdType)
	return fmt.Sprintf(`<div class="banner-strip">
  <a href="/api/r/%s/click/%s" target="_blank" rel="noopener">%s</a>
</div>`, slug, adID, media)
}

func bannerPlaceholder() string {
	return `<div class="banner-strip" style="background:#f8fafc;border:1px dashed #cbd5e1;display:flex;align-items:center;justify-content:center;padding:20px;text-align:center;border-radius:14px;box-shadow:none">
  <a href="/" style="color:#64748b;font-size:13px;font-weight:500;text-decoration:none;font-family:'DM Sans',sans-serif" target="_blank" rel="noopener">Advertise Here</a>
</div>`
}

func nativeCard(ad *repository.Ad, slug string) string {
	adID := ad.ID.String()
	media := adMediaTag(ad.ImageUrl, ad.AdType)
	desc := ""
	if ad.Description.Valid {
		desc = ad.Description.String
	}
	if desc == "" {
		desc = "Sponsored content"
	}
	return fmt.Sprintf(`<div class="native-card">
  <div class="media">%s</div>
  <div class="body">
    <div class="title">%s</div>
    <div class="desc">%s</div>
    <a href="/api/r/%s/click/%s" class="btn" target="_blank" rel="noopener">Learn More</a>
  </div>
</div>`, media, ad.Title, desc, slug, adID)
}

func nativePlaceholder() string {
	return `<div class="native-card" style="background:#f8fafc;border:1px dashed #cbd5e1;display:flex;align-items:center;justify-content:center;min-height:200px;text-align:center;box-shadow:none">
  <div>
    <div style="font-size:14px;font-weight:600;color:#64748b;margin-bottom:12px;font-family:'DM Sans',sans-serif">Advertise Here</div>
    <a href="/" style="display:inline-block;padding:8px 16px;border-radius:8px;font-size:12px;font-weight:600;text-decoration:none;background:linear-gradient(135deg,#4b5563,#111827);color:#fff;box-shadow:0 4px 12px rgba(0,0,0,.22);font-family:'DM Sans',sans-serif" target="_blank" rel="noopener">Advertise Here</a>
  </div>
</div>`
}

func videoSection(ad *repository.Ad, slug string) string {
	adID := ad.ID.String()
	return fmt.Sprintf(`<div class="video-section">
  <video src="%s" autoplay muted loop playsinline webkit-playsinline></video>
  <div class="body">
    <div class="title">%s</div>
    <a href="/api/r/%s/click/%s" class="btn" target="_blank" rel="noopener">Learn More</a>
  </div>
</div>`, ad.ImageUrl, ad.Title, slug, adID)
}

func videoPlaceholder() string {
	return `<div class="video-section" style="background:#f8fafc;border:1px dashed #cbd5e1;display:flex;align-items:center;justify-content:center;min-height:200px;text-align:center;box-shadow:none">
  <div>
    <div style="font-size:14px;font-weight:600;color:#64748b;margin-bottom:12px;font-family:'DM Sans',sans-serif">Advertise Here</div>
    <a href="/" style="display:inline-block;padding:8px 16px;border-radius:8px;font-size:12px;font-weight:600;text-decoration:none;background:linear-gradient(135deg,#4b5563,#111827);color:#fff;box-shadow:0 4px 12px rgba(0,0,0,.22);font-family:'DM Sans',sans-serif" target="_blank" rel="noopener">Advertise Here</a>
  </div>
</div>`
}

func interstSection(ad *repository.Ad, slug string) string {
	adID := ad.ID.String()
	media := adMediaTag(ad.ImageUrl, ad.AdType)
	desc := ""
	if ad.Description.Valid {
		desc = ad.Description.String
	}
	if desc == "" {
		desc = "Sponsored"
	}
	return fmt.Sprintf(`<div class="interst-section">
  %s
  <div class="overlay">
    <div class="title">%s</div>
    <div class="desc">%s</div>
    <a href="/api/r/%s/click/%s" class="btn" target="_blank" rel="noopener">Learn More</a>
  </div>
</div>`, media, ad.Title, desc, slug, adID)
}

func interstPlaceholder() string {
	return `<div class="interst-section" style="background:#f8fafc;border:1px dashed #cbd5e1;display:flex;align-items:center;justify-content:center;min-height:250px;text-align:center;box-shadow:none">
  <div>
    <div style="font-size:18px;font-weight:600;color:#64748b;margin-bottom:12px;font-family:'DM Sans',sans-serif">Advertise Here</div>
    <a href="/" style="display:inline-block;padding:10px 24px;border-radius:10px;font-size:14px;font-weight:600;text-decoration:none;background:linear-gradient(135deg,#4b5563,#111827);color:#fff;box-shadow:0 4px 12px rgba(0,0,0,.22);font-family:'DM Sans',sans-serif" target="_blank" rel="noopener">Advertise Here</a>
  </div>
</div>`
}
