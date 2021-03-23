# urort-downloader
Fetch metadata and MP3s from NRK urørt (http://urort.p3.no/)

Work-in-progress (of course : )
Restarted in 2021 because they changed their API. The old version is in another branch

## What
NRK Urørt is new music from Norway. This code fetches MP3's from the currated list of recommended songs. 

## Why
I want to:
 - play the songs locally
 - keep the mp3's I like
 - have software remember what I have already downloaded


## How
 - Fetch metadata from the Urort API
 - Refresh the local database (badger, stored in files)
 - Get and store any missing MP3s from the Urort servers
