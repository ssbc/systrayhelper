# systrayhelper [![Build Status](https://travis-ci.org/ssbc/systrayhelper.svg?branch=master)](https://travis-ci.org/ssbc/systrayhelper)
A portable version of [go systray](https://github.com/getlantern/systray), using json objects over stdio to communicate with other languages.

Note(cryptix): this is the cleanup fork of [forked-systrayhelper](https://github.com/ssbc/forked-systrayhelper) sorry for the mess..

## Installation

Install [Go](https://golang.org) and run `go get -v github.com/ssbc/systrayhelper`. The binary will be in `$PATH/go/bin` by default.

Or clone this repo and build it:

```bash
git clone https://github.com/ssbc/systrayhelper
cd systrayhelper
go build
./systrayhelper --test
```

**Linux**: As noted on the [getlantern/systray repo](https://github.com/getlantern/systray#linux), you need to install these two packages: `libgtk-3-dev` and `libappindicator3-dev`

On windows and darwin it should built just fine.


## Protocol

tray binary =>  
=> ready  `{"type": "ready"}`  
<= init menu
```json
{
  "icon": "<base64 string of image>",
  "title": "Title",
  "tooltip": "Tooltips",
  "items":[{
    "title": "aa",
    "tooltip":"bb",
    "checked": true,
    "enabled": true
  }, {
    "title": "aa2",
    "tooltip":"bb",
    "checked": false,
    "enabled": true
  }]}
```
=> clicked  
```json
{
  "type":"clicked",
  "item":{"title":"aa","tooltip":"bb","enabled":true,"checked":true},
  "menu":{"icon":"","title":"","tooltip":"","items":null},
  "seq_id":0
}
```
<= update-item / update-menu / update-item-and-menu
```json
{
  "type": "update-item",
  "item": {"title":"aa3","tooltip":"bb","enabled":true,"checked":true},
  "seq_id": 0
}
```

## Repo Init

Had to start somewhere, so I took [99b200...](https://github.com/ssbc/forked-systray/commit/99b2002b2e34f6381a04f365907f2e9dcd8837ea) from the previous repo.


```bash
$ git clone https://github.com/ssbc/forked-systray small
$ cd small && git reset --hard 99b2002b2e34f6381a04f365907f2e9dcd8837ea
HEAD is now at 99b2002 changed name to forked-systray
$ archive=../systrayhelper-new.tar
$ tar cf $archive * && xz $archive
$ ls -sh $archive.xz && sbot blobs.add < $archive.xz
8.5K ../systrayhelper-new.tar.xz
&YmiqTDdWgNzdAczEo+DNHKb1X1X4hyHNrWc7rFgIW84=.sha256
```
