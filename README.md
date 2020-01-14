# Flash Torrent
A CLI interface for the flash torrent eco-system. Designed to streamline searching, torrenting, and watching videos.

## Example
![image](https://raw.githubusercontent.com/bcopp/flash-torrent-cli/master/example.png)

## Quick Start
- Install Go & Go Dependencies
```
// Install Go from: https://golang.org/doc/install
go get -v github.com/PuerkitoBio/goquery
go get -v github.com/Sirupsen/logrus
go get -v github.com/jessevdk/go-flags
```
- Install Torrent Dependencies (Ubuntu)
```
sudo add-apt-repository ppa:transmissionbt/ppa
sudo apt-get update
sudo apt install transmission-cli transmission-common transmission-daemon snap
sudo apt install vlc
```
- Compile main.go and vlc/_helper
- Run the Program
`./main -f $HOME/Downloads`
or
`./main -f $HOME/Downloads -s "my search"`

## Options
```
Usage:
  main [OPTIONS]

Application Options:
  -v, --verbose          Show verbose debug information
  -m, --meme             Start program with a different name each time
  -s, --search=SEARCH    The search you would like to be made on "The Pirate Bay"
  -f, --folder=FOLDER    Folder to download torrents to. DEFAULTS to $HOME/Downloads
  -n, --no-vlc           Do NOT run VLC when the file has finished downloading
  -j, --json             Print the results to the screen as JSON

Help Options:
  -h, --help             Show this help message
```
