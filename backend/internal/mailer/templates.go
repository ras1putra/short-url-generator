package mailer

const verificationTemplate = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background-color:#0A0A0A;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
<table width="100%" cellpadding="0" cellspacing="0" style="background-color:#0A0A0A;padding:40px 0">
<tr><td align="center">
<table width="480" cellpadding="0" cellspacing="0" style="background-color:#111111;border-radius:16px;overflow:hidden;border:1px solid rgba(255,255,255,0.08)">
<tr><td style="padding:40px 32px 32px;text-align:center">
<table style="margin:0 auto 24px">
<tr><td style="background-color:#6EE7B7;width:36px;height:36px;border-radius:8px;text-align:center;vertical-align:middle;font-size:16px;line-height:36px;color:#0A0A0A;font-weight:bold">&#8646;</td></tr>
</table>
<h1 style="margin:0 0 4px;font-size:22px;color:#ffffff;font-weight:800;letter-spacing:-0.02em">go-short</h1>
<p style="margin:0 0 24px;font-size:13px;color:rgba(255,255,255,0.4);font-weight:500">VERIFY YOUR EMAIL</p>
<p style="margin:0 0 24px;font-size:15px;color:rgba(255,255,255,0.7);line-height:1.6">Hi {{.Name}},<br/><br/>Click the button below to verify your email address and activate your account.</p>
<a href="{{.Link}}" style="display:inline-block;padding:14px 36px;background-color:#6EE7B7;color:#0A0A0A;text-decoration:none;border-radius:10px;font-size:14px;font-weight:700;letter-spacing:0.02em">Verify Email</a>
<p style="margin:24px 0 0;font-size:12px;color:rgba(255,255,255,0.3)">This link expires in 10 minutes.</p>
</td></tr>
<tr><td style="padding:20px 32px;text-align:center;border-top:1px solid rgba(255,255,255,0.06);background-color:rgba(255,255,255,0.02)">
<p style="margin:0;font-size:11px;color:rgba(255,255,255,0.25)">go-short &mdash; Lightning-fast redirects</p>
</td></tr>
</table>
</td></tr>
</table>
</body>
</html>`

const passwordResetTemplate = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background-color:#0A0A0A;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
<table width="100%" cellpadding="0" cellspacing="0" style="background-color:#0A0A0A;padding:40px 0">
<tr><td align="center">
<table width="480" cellpadding="0" cellspacing="0" style="background-color:#111111;border-radius:16px;overflow:hidden;border:1px solid rgba(255,255,255,0.08)">
<tr><td style="padding:40px 32px 32px;text-align:center">
<table style="margin:0 auto 24px">
<tr><td style="background-color:#6EE7B7;width:36px;height:36px;border-radius:8px;text-align:center;vertical-align:middle;font-size:16px;line-height:36px;color:#0A0A0A;font-weight:bold">&#8646;</td></tr>
</table>
<h1 style="margin:0 0 4px;font-size:22px;color:#ffffff;font-weight:800;letter-spacing:-0.02em">go-short</h1>
<p style="margin:0 0 24px;font-size:13px;color:rgba(255,255,255,0.4);font-weight:500">RESET YOUR PASSWORD</p>
<p style="margin:0 0 24px;font-size:15px;color:rgba(255,255,255,0.7);line-height:1.6">Hi {{.Name}},<br/><br/>Click the button below to reset your password. If you didn't request this, you can safely ignore this email.</p>
<a href="{{.Link}}" style="display:inline-block;padding:14px 36px;background-color:#6EE7B7;color:#0A0A0A;text-decoration:none;border-radius:10px;font-size:14px;font-weight:700;letter-spacing:0.02em">Reset Password</a>
<p style="margin:24px 0 0;font-size:12px;color:rgba(255,255,255,0.3)">This link expires in 10 minutes.</p>
</td></tr>
<tr><td style="padding:20px 32px;text-align:center;border-top:1px solid rgba(255,255,255,0.06);background-color:rgba(255,255,255,0.02)">
<p style="margin:0;font-size:11px;color:rgba(255,255,255,0.25)">go-short &mdash; Lightning-fast redirects</p>
</td></tr>
</table>
</td></tr>
</table>
</body>
</html>`
