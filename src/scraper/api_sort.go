package scraper

type apiSort struct{ answers, comments, order string }

var apiSorts = map[string]apiSort{
	"votes":    {"votes", "votes", "desc"},
	"trending": {"votes", "votes", "desc"},
	"newest":   {"activity", "creation", "desc"},
	"oldest":   {"creation", "creation", "asc"},
}

func apiSortFor(sortValue string) apiSort {
	if s, ok := apiSorts[sortValue]; ok {
		return s
	}
	return apiSorts["votes"]
}
