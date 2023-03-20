package scraper

type Item struct {
	ProductID       string `json:"product_id"`
	Price           string `json:"price"`
	Title           string `json:"title"`
	HREF            string `json:"href"`
	Desc            string `json:"desc"`
	GameTime        string `json:"gameTime"`
	NumberOfPlayers string `json:"numberOfPlayers"`
	Age             string `json:"age"`
	SrcPageRef      string `json:"srcPageRef"`
}
