$ErrorActionPreference = "Stop"
$B = "http://localhost:8080/api"
$script:fail = 0
function Assert($name, $cond, $detail) {
  if ($cond) { Write-Host "PASS  $name" }
  else { Write-Host "FAIL  $name  $detail"; $script:fail++ }
}
function Call($path, $token, $bodyObj, $method) {
  $headers = @{ }
  if ($token) { $headers["Authorization"] = "Bearer $token" }
  try {
    if ($null -ne $bodyObj) {
      $json = $bodyObj | ConvertTo-Json -Depth 6
      if (-not $method) { $method = "Post" }
      return Invoke-RestMethod -Uri "$B$path" -Method $method -Headers $headers -ContentType "application/json; charset=utf-8" -Body ([System.Text.Encoding]::UTF8.GetBytes($json))
    } elseif ($method -eq "Post") {
      return Invoke-RestMethod -Uri "$B$path" -Method Post -Headers $headers -ContentType "application/json; charset=utf-8" -Body ([System.Text.Encoding]::UTF8.GetBytes("{}"))
    } else {
      return Invoke-RestMethod -Uri "$B$path" -Method Get -Headers $headers
    }
  } catch {
    $resp = $_.Exception.Response
    $code = if ($resp) { [int]$resp.StatusCode } else { 0 }
    $msg = ""
    if ($resp) {
      $sr = New-Object System.IO.StreamReader($resp.GetResponseStream())
      try { $msg = ($sr.ReadToEnd() | ConvertFrom-Json).error } catch {}
    }
    return @{ __error = $true; code = $code; error = $msg }
  }
}

# --- seed state: ali 70/0  bo 110/0  ci 50/0
$ali = Call "/login" $null @{ username="ali"; password="123456" }
$bo  = Call "/login" $null @{ username="bo";  password="123456" }
$ci  = Call "/login" $null @{ username="ci";  password="123456" }
$ta=$ali.token; $tb=$bo.token; $tc=$ci.token
Assert "seed ali 70/0"   ($ali.user.coins -eq 70  -and $ali.user.frozen_coins -eq 0) "$($ali.user.coins)/$($ali.user.frozen_coins)"
Assert "seed bo 110"     ($bo.user.coins  -eq 110) "$($bo.user.coins)"
Assert "seed ci 50"      ($ci.user.coins  -eq 50)  "$($ci.user.coins)"

# --- register gift
$nu = Call "/register" $null @{ username="newbie01"; password="123456"; nickname="Newbie"; avatar_emoji="dumpling" }
Assert "register gift 100" ($nu.user.coins -eq 100 -and $nu.user.frozen_coins -eq 0) "$($nu.user.coins)"
$dup = Call "/register" $null @{ username="ali"; password="123456"; nickname="x" }
Assert "dup username rejected" ($dup.__error -and $dup.code -eq 409) "code=$($dup.code)"

# --- create wish freeze 25
$w = Call "/wishes" $ta @{ restaurant_name="Tofu Shop"; address="Test Rd"; dish_list="salty,sweet"; reason="which wins"; reward_coins=25 }
$wid = $w.wish.id
$me = Call "/me" $ta
Assert "freeze 25 -> 45/25" ($me.user.coins -eq 45 -and $me.user.frozen_coins -eq 25) "$($me.user.coins)/$($me.user.frozen_coins)"
Assert "wish reward/frozen 25" ($w.wish.reward_coins -eq 25 -and $w.wish.status -eq "open") "$($w.wish.reward_coins)/$($w.wish.status)"

# bad inputs
$bad0 = Call "/wishes" $ta @{ restaurant_name="X"; address="X"; dish_list="X"; reason="X"; reward_coins=0 }
Assert "reward 0 rejected" ($bad0.__error -and $bad0.code -eq 400) "code=$($bad0.code)"
$over = Call "/wishes" $ta @{ restaurant_name="X"; address="X"; dish_list="X"; reason="X"; reward_coins=99999 }
Assert "overspend rejected 400" ($over.__error -and $over.code -eq 400) "code=$($over.code)"
$me2 = Call "/me" $ta
Assert "failed freeze changed nothing 45/25" ($me2.user.coins -eq 45 -and $me2.user.frozen_coins -eq 25) "$($me2.user.coins)/$($me2.user.frozen_coins)"

# --- append 10
$w2 = Call "/wishes/$wid/append" $ta @{ amount=10 }
$me = Call "/me" $ta
Assert "append 10 -> 35/35" ($me.user.coins -eq 35 -and $me.user.frozen_coins -eq 35 -and $w2.wish.reward_coins -eq 35) "u=$($me.user.coins)/$($me.user.frozen_coins) r=$($w2.wish.reward_coins)"
$apOver = Call "/wishes/$wid/append" $ta @{ amount=9999 }
Assert "append overspend rejected" ($apOver.__error -and $apOver.code -eq 400) "code=$($apOver.code)"

# others cannot append bo
$boAppend = Call "/wishes/$wid/append" $tb @{ amount=5 }
Assert "non-owner append rejected 404" ($boAppend.__error -and $boAppend.code -eq 404) "code=$($boAppend.code)"

# --- taster submits
$r = Call "/wishes/$wid/responses" $tb @{ photos=@("/uploads/seed/cake.svg"); comment="Salty silky with shrimp, sweet osmanthus syrup, loved both!"; rating=5 }
$rid = $r.response.id
Assert "taster response created" (-not $r.__error) "$($r.error)"
$dupR = Call "/wishes/$wid/responses" $tb @{ photos=@("/uploads/seed/cake.svg"); comment="duplicate tasting should be rejected now"; rating=4 }
Assert "duplicate taste 400" ($dupR.__error -and $dupR.code -eq 400) "code=$($dupR.code)"
$selfR = Call "/wishes/$wid/responses" $ta @{ photos=@("/uploads/seed/cake.svg"); comment="self taste should be rejected right now"; rating=3 }
Assert "self taste 400" ($selfR.__error -and $selfR.code -eq 400) "code=$($selfR.code)"

# ci also responds
$r2 = Call "/wishes/$wid/responses" $tc @{ photos=@("/uploads/seed/bbq.svg"); comment="I preferred the sweet one, soft tofu and fragrant syrup!"; rating=4 }
$rid2 = $r2.response.id

# non-owner cannot settle
$boSet = Call "/wishes/$wid/settle" $tb @{ response_id=$rid }
Assert "non-owner settle rejected 404" ($boSet.__error -and $boSet.code -eq 404) "code=$($boSet.code)"

# --- owner settles choosing bo (reward 35)
$ok = Call "/wishes/$wid/settle" $ta @{ response_id=$rid }
Assert "settle ok" (-not $ok.__error) "$($ok.error)"
$ma = Call "/me" $ta; $mb = Call "/me" $tb; $mc = Call "/me" $tc
Assert "after settle ali 35/0" ($ma.user.coins -eq 35 -and $ma.user.frozen_coins -eq 0) "$($ma.user.coins)/$($ma.user.frozen_coins)"
# bo 自身在种子数据中发过 20 币悬赏（w2 仍 open），故冻结 20 一直挂着
Assert "after settle bo 145/20" ($mb.user.coins -eq 145 -and $mb.user.frozen_coins -eq 20) "$($mb.user.coins)/$($mb.user.frozen_coins)"
Assert "ci untouched 50" ($mc.user.coins -eq 50) "$($mc.user.coins)"

$again = Call "/wishes/$wid/settle" $ta @{ response_id=$rid2 }
Assert "double settle 409" ($again.__error -and $again.code -eq 409) "code=$($again.code)"

$detail = Call "/wishes/$wid"
$adopted = @{}; foreach ($x in $detail.wish.responses) { $adopted[$x.user.username] = $x.is_adopted }
Assert "wish settled, only bo adopted" ($detail.wish.status -eq "settled" -and $adopted["bo"] -eq $true -and $adopted["ci"] -eq $false) "$($detail.wish.status) $($adopted | Out-String)"

# no new responses after settled
$locked = Call "/wishes/$wid/responses" $tc @{ photos=@("/uploads/x.svg"); comment="cannot taste a settled wish anymore here"; rating=2 }
Assert "taste settled wish 409" ($locked.__error -and $locked.code -eq 409) "code=$($locked.code)"

# --- ledger identity: sum(amount) == coins for each user
foreach ($pair in @(@($ta,"ali",35), @($tb,"bo",145), @($tc,"ci",50), @($nu.token,"newbie",100))) {
  $t = Call "/my/transactions" $pair[0]
  $sum = ($t.items | Measure-Object -Property amount -Sum).Sum
  Assert "ledger identity $($pair[1]) sum=$sum bal=$($pair[2])" ($sum -eq $pair[2]) "sum=$sum"
}

# --- cancellation refund path (ali has 35 available now)
$w3 = Call "/wishes" $ta @{ restaurant_name="Milk Tea"; address="corner"; dish_list="lemon tea"; reason="refund path"; reward_coins=12 }
$pre = (Call "/me" $ta).user.coins
$cn = Call "/wishes/$($w3.wish.id)/cancel" $ta $null "Post"
$post = Call "/me" $ta
Assert "cancel: 23 freeze -> back to 35/0" ($pre -eq 23 -and $post.user.coins -eq 35 -and $post.user.frozen_coins -eq 0) "$pre -> $($post.user.coins)/$($post.user.frozen_coins)"
$cn2 = Call "/wishes/$($w3.wish.id)/cancel" $ta $null "Post"
Assert "double cancel 409" ($cn2.__error -and $cn2.code -eq 409) "code=$($cn2.code)"

# --- feed + my lists
$feed = Call "/wishes?status=open"
Assert "feed open only" (($feed.items | Where-Object { $_.status -ne "open" }).Count -eq 0) ""
$myw = Call "/my/wishes" $ta
Assert "my wishes count>=3" ($myw.items.Count -ge 3) "$($myw.items.Count)"
$myr = Call "/my/responses" $tb
Assert "bo my-responses has adopted one" (($myr.items | Where-Object { $_.is_adopted -eq $true }).Count -ge 1) ""

# --- auth guard
try {
  Invoke-RestMethod -Uri "$B/me" -Method Get
  Assert "no-token /me rejected" $false ""
} catch {
  Assert "no-token /me rejected 401" ([int]$_.Exception.Response.StatusCode -eq 401) ""
}

Write-Host ""
if ($script:fail -eq 0) { Write-Host "ALL TESTS PASSED" } else { Write-Host "$script:fail TEST(S) FAILED"; exit 1 }
