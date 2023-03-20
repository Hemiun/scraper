# Scrapper

Scraping hobbygames shop catalog into csv file. 

## Features
- ProductID
- Price
- Title
- Reference to item card
- Description
- Game time
- Number of players
- Age

## How to Use
- Build executable file. For example
`go build -o scraper`
- Run new scraping session. 
`./scraper start`  
For every session a new data folder will be created (`./data/yyyy-mm-dd_hh24miss`) 
- The final output will be stored in a CSV file (result.csv) in session data folder.
- Available Commands:
```
  clean       Clean data folder
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  start       A brief description of your command
```




> **Disclaimer**<a name="disclaimer" />: Please note that this is a research project. I am by no means responsible for any usage of this tool. Use it on your behalf. I am not responsible for any damages, this scripts and tools only for testing purpose. Everything here in this repository has been made for  demonstration purposes only, testing environment only. Thanks.