# go-flickr-archive

Don't try to use this yet. It won't do what you expect and almost certainly has bugs.

## Tools

### flickr-archive

This _does not_ archive private photos yet. Or the actual photos, for that matter...

```
$> ./bin/flickr-archive -api-key API-KEY -root . -username bees

2017/12/04 10:09:40 ARCHIVE bees/public/2009/03/17/3364783370/3364783370_f54d98e5fa_i.json
2017/12/04 10:09:41 ARCHIVE bees/public/2009/03/17/3363961741/3363961741_84c3033dce_i.json
2017/12/04 10:09:42 ARCHIVE bees/public/2009/03/17/3364783168/3364783168_1dcbf996db_i.json

$> cat bees/public/2009/03/17/3364783370/3364783370_f54d98e5fa_i.json
{"photo":{"id":"3364783370","secret":"f54d98e5fa","server":"3549","farm":4,"dateuploaded":"1237348574","isfavorite":0,"license":"0","safety_level":"0","rotation":0,"originalsecret":"9a1a99b776","originalformat":"jpg","owner":{"nsid":"36544748@N08","username":"BEES","realname":"BEES","location":"","iconserver":"3601","iconfarm":4,"path_alias":"beesoficial"},"title":{"_content":"\u00d3ia n\u00f3is"},"description":{"_content":""},"visibility":{"ispublic":1,"isfriend":0,"isfamily":0},"dates":{"posted":"1237348574","taken":"2009-03-17 20:56:14","takengranularity":"0","takenunknown":0,"lastupdate":"1237416538"},"views":"12","editability":{"cancomment":0,"canaddmeta":0},"publiceditability":{"cancomment":1,"canaddmeta":0},"usage":{"candownload":1,"canblog":0,"canprint":0,"canshare":1},"comments":{"_content":"0"},"notes":{"note":[]},"people":{"haspeople":0},"tags":{"tag":[]},"urls":{"url":[{"type":"photopage","_content":"https:\/\/www.flickr.com\/photos\/beesoficial\/3364783370\/"}]},"media":"photo"},"stat":"ok"}
```
